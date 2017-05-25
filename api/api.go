package api

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"

	"github.com/sirupsen/logrus"

	"math/big"
	"net"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

const (
	jsonMediaType = "application/json"
	textMediaType = "text/plain"
)

var userAgentPattern = regexp.MustCompile(
	`^(?:curl|Wget|fetch\slibfetch|ddclient|Go-http-client|HTTPie)\/.*|Go\s1\.1\spackage\shttp$`,
)

type API struct {
	Template string
	IPHeader string
	oracle   Oracle
	log      *logrus.Logger
}

type Response struct {
	IP        net.IP   `json:"ip"`
	IPDecimal *big.Int `json:"ip_decimal"`
	Country   string   `json:"country,omitempty"`
	City      string   `json:"city,omitempty"`
	Hostname  string   `json:"hostname,omitempty"`
}

type PortResponse struct {
	IP        net.IP `json:"ip"`
	Port      uint64 `json:"port"`
	Reachable bool   `json:"reachable"`
}

func New(oracle Oracle, logger *logrus.Logger) *API {
	return &API{oracle: oracle, log: logger}
}

func ipToDecimal(ip net.IP) *big.Int {
	i := big.NewInt(0)
	if to4 := ip.To4(); to4 != nil {
		i.SetBytes(to4)
	} else {
		i.SetBytes(ip)
	}
	return i
}

func ipFromRequest(header string, r *http.Request) (net.IP, error) {
	remoteIP := r.Header.Get(header)
	if remoteIP == "" {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			return nil, err
		}
		remoteIP = host
	}
	ip := net.ParseIP(remoteIP)
	if ip == nil {
		return nil, fmt.Errorf("could not parse IP: %s", remoteIP)
	}
	return ip, nil
}

func (a *API) newResponse(r *http.Request) (Response, error) {
	ip, err := ipFromRequest(a.IPHeader, r)
	if err != nil {
		return Response{}, err
	}
	ipDecimal := ipToDecimal(ip)
	country, err := a.oracle.LookupCountry(ip)
	if err != nil {
		a.log.Debug(err)
	}
	city, err := a.oracle.LookupCity(ip)
	if err != nil {
		a.log.Debug(err)
	}
	hostnames, err := a.oracle.LookupAddr(ip)
	if err != nil {
		a.log.Debug(err)
	}
	return Response{
		IP:        ip,
		IPDecimal: ipDecimal,
		Country:   country,
		City:      city,
		Hostname:  strings.Join(hostnames, " "),
	}, nil
}

func (a *API) newPortResponse(r *http.Request) (PortResponse, error) {
	vars := mux.Vars(r)
	port, err := strconv.ParseUint(vars["port"], 10, 16)
	if err != nil {
		return PortResponse{Port: port}, err
	}
	if port < 1 || port > 65355 {
		return PortResponse{Port: port}, fmt.Errorf("invalid port: %d", port)
	}
	ip, err := ipFromRequest(a.IPHeader, r)
	if err != nil {
		return PortResponse{Port: port}, err
	}
	err = a.oracle.LookupPort(ip, port)
	return PortResponse{
		IP:        ip,
		Port:      port,
		Reachable: err == nil,
	}, nil
}

func (a *API) CLIHandler(w http.ResponseWriter, r *http.Request) *appError {
	ip, err := ipFromRequest(a.IPHeader, r)
	if err != nil {
		return internalServerError(err)
	}
	io.WriteString(w, ip.String()+"\n")
	return nil
}

func (a *API) CLICountryHandler(w http.ResponseWriter, r *http.Request) *appError {
	response, err := a.newResponse(r)
	if err != nil {
		return internalServerError(err)
	}
	io.WriteString(w, response.Country+"\n")
	return nil
}

func (a *API) CLICityHandler(w http.ResponseWriter, r *http.Request) *appError {
	response, err := a.newResponse(r)
	if err != nil {
		return internalServerError(err)
	}
	io.WriteString(w, response.City+"\n")
	return nil
}

func (a *API) JSONHandler(w http.ResponseWriter, r *http.Request) *appError {
	response, err := a.newResponse(r)
	if err != nil {
		return internalServerError(err).AsJSON()
	}
	b, err := json.Marshal(response)
	if err != nil {
		return internalServerError(err).AsJSON()
	}
	w.Header().Set("Content-Type", jsonMediaType)
	w.Write(b)
	return nil
}

func (a *API) PortHandler(w http.ResponseWriter, r *http.Request) *appError {
	response, err := a.newPortResponse(r)
	if err != nil {
		return badRequest(err).WithMessage(fmt.Sprintf("Invalid port: %d", response.Port)).AsJSON()
	}
	b, err := json.Marshal(response)
	if err != nil {
		return internalServerError(err).AsJSON()
	}
	w.Header().Set("Content-Type", jsonMediaType)
	w.Write(b)
	return nil
}

func (a *API) DefaultHandler(w http.ResponseWriter, r *http.Request) *appError {
	response, err := a.newResponse(r)
	if err != nil {
		return internalServerError(err)
	}
	t, err := template.New(filepath.Base(a.Template)).ParseFiles(a.Template)
	if err != nil {
		return internalServerError(err)
	}
	var data = struct {
		Response
		Oracle
	}{response, a.oracle}
	if err := t.Execute(w, &data); err != nil {
		return internalServerError(err)
	}
	return nil
}

func (a *API) NotFoundHandler(w http.ResponseWriter, r *http.Request) *appError {
	err := notFound(nil).WithMessage("404 page not found")
	if r.Header.Get("accept") == jsonMediaType {
		err = err.AsJSON()
	}
	return err
}

func cliMatcher(r *http.Request, rm *mux.RouteMatch) bool {
	return userAgentPattern.MatchString(r.UserAgent())
}

type appHandler func(http.ResponseWriter, *http.Request) *appError

func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if e := fn(w, r); e != nil { // e is *appError
		// When Content-Type for error is JSON, we need to marshal the response into JSON
		if e.IsJSON() {
			var data = struct {
				Error string `json:"error"`
			}{e.Message}
			b, err := json.Marshal(data)
			if err != nil {
				panic(err)
			}
			e.Message = string(b)
		}
		// Set Content-Type of response if set in error
		if e.ContentType != "" {
			w.Header().Set("Content-Type", e.ContentType)
		}
		w.WriteHeader(e.Code)
		io.WriteString(w, e.Message)
	}
}

func (a *API) Router() http.Handler {
	r := mux.NewRouter()

	// JSON
	r.Handle("/", appHandler(a.JSONHandler)).Methods("GET").Headers("Accept", jsonMediaType)
	r.Handle("/json", appHandler(a.JSONHandler)).Methods("GET")

	// CLI
	r.Handle("/", appHandler(a.CLIHandler)).Methods("GET").MatcherFunc(cliMatcher)
	r.Handle("/", appHandler(a.CLIHandler)).Methods("GET").Headers("Accept", textMediaType)
	r.Handle("/ip", appHandler(a.CLIHandler)).Methods("GET")
	r.Handle("/country", appHandler(a.CLICountryHandler)).Methods("GET")
	r.Handle("/city", appHandler(a.CLICityHandler)).Methods("GET")

	// Browser
	r.Handle("/", appHandler(a.DefaultHandler)).Methods("GET")

	// Port testing
	r.Handle("/port/{port:[0-9]+}", appHandler(a.PortHandler)).Methods("GET")

	// Not found handler which returns JSON when appropriate
	r.NotFoundHandler = appHandler(a.NotFoundHandler)

	return r
}
