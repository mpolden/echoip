package api

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

const APPLICATION_JSON = "application/json"

var USER_AGENT_RE = regexp.MustCompile(
	`^(?:curl|Wget|fetch\slibfetch|Go-http-client|HTTPie)\/.*|Go\s1\.1\spackage\shttp$`,
)

type API struct {
	Template string
	IPHeader string
	oracle   Oracle
}

type Response struct {
	IP       net.IP `json:"ip"`
	Country  string `json:"country,omitempty"`
	City     string `json:"city,omitempty"`
	Hostname string `json:"hostname,omitempty"`
}

type TestPortResponse struct {
	IP        net.IP `json:"ip"`
	Port      uint64 `json:"port"`
	Reachable bool   `json:"reachable"`
}

func New(oracle Oracle) *API {
	return &API{oracle: oracle}
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
	country, err := a.oracle.LookupCountry(ip)
	if err != nil {
		log.Print(err)
	}
	city, err := a.oracle.LookupCity(ip)
	if err != nil {
		log.Print(err)
	}
	hostnames, err := a.oracle.LookupAddr(ip.String())
	if err != nil {
		log.Print(err)
	}
	return Response{
		IP:       ip,
		Country:  country,
		City:     city,
		Hostname: strings.Join(hostnames, " "),
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
	w.Header().Set("Content-Type", APPLICATION_JSON)
	w.Write(b)
	return nil
}

func (a *API) PortHandler(w http.ResponseWriter, r *http.Request) *appError {
	vars := mux.Vars(r)
	port, err := strconv.ParseUint(vars["port"], 10, 16)
	if err != nil {
		return badRequest(err).WithMessage("Invalid port: " + vars["port"]).AsJSON()
	}
	if port < 1 || port > 65355 {
		return badRequest(nil).WithMessage("Invalid port: " + vars["port"]).AsJSON()
	}
	ip, err := ipFromRequest(a.IPHeader, r)
	if err != nil {
		return internalServerError(err).AsJSON()
	}
	err = a.oracle.LookupPort(ip, port)
	response := TestPortResponse{
		IP:        ip,
		Port:      port,
		Reachable: err == nil,
	}
	b, err := json.Marshal(response)
	if err != nil {
		return internalServerError(err).AsJSON()
	}
	w.Header().Set("Content-Type", APPLICATION_JSON)
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
	if r.Header.Get("accept") == APPLICATION_JSON {
		err = err.AsJSON()
	}
	return err
}

func cliMatcher(r *http.Request, rm *mux.RouteMatch) bool {
	return USER_AGENT_RE.MatchString(r.UserAgent())
}

type appHandler func(http.ResponseWriter, *http.Request) *appError

func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if e := fn(w, r); e != nil { // e is *appError
		if e.Error != nil {
			log.Print(e.Error)
		}
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

func (a *API) Handlers() http.Handler {
	r := mux.NewRouter()

	// JSON
	r.Handle("/", appHandler(a.JSONHandler)).Methods("GET").Headers("Accept", APPLICATION_JSON)
	r.Handle("/json", appHandler(a.JSONHandler)).Methods("GET")

	// CLI
	r.Handle("/", appHandler(a.CLIHandler)).Methods("GET").MatcherFunc(cliMatcher)
	r.Handle("/ip", appHandler(a.CLIHandler)).Methods("GET").MatcherFunc(cliMatcher)
	r.Handle("/country", appHandler(a.CLICountryHandler)).Methods("GET").MatcherFunc(cliMatcher)
	r.Handle("/city", appHandler(a.CLICityHandler)).Methods("GET").MatcherFunc(cliMatcher)

	// Browser
	r.Handle("/", appHandler(a.DefaultHandler)).Methods("GET")

	// Port testing
	r.Handle("/port/{port:[0-9]+}", appHandler(a.PortHandler)).Methods("GET")

	// Not found handler which returns JSON when appropriate
	r.NotFoundHandler = appHandler(a.NotFoundHandler)

	return r
}

func (a *API) ListenAndServe(addr string) error {
	http.Handle("/", a.Handlers())
	return http.ListenAndServe(addr, nil)
}
