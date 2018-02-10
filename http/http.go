package http

import (
	"encoding/json"
	"fmt"
	"html/template"

	"github.com/mpolden/ipd/useragent"
	"github.com/sirupsen/logrus"

	"math/big"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

const (
	jsonMediaType = "application/json"
	textMediaType = "text/plain"
)

type Server struct {
	Template string
	IPHeader string
	oracle   Oracle
	log      *logrus.Logger
}

type Response struct {
	IP         net.IP   `json:"ip"`
	IPDecimal  *big.Int `json:"ip_decimal"`
	Country    string   `json:"country,omitempty"`
	CountryISO string   `json:"country_iso,omitempty"`
	City       string   `json:"city,omitempty"`
	Hostname   string   `json:"hostname,omitempty"`
}

type PortResponse struct {
	IP        net.IP `json:"ip"`
	Port      uint64 `json:"port"`
	Reachable bool   `json:"reachable"`
}

func New(oracle Oracle, logger *logrus.Logger) *Server {
	return &Server{oracle: oracle, log: logger}
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

func (s *Server) newResponse(r *http.Request) (Response, error) {
	ip, err := ipFromRequest(s.IPHeader, r)
	if err != nil {
		return Response{}, err
	}
	ipDecimal := ipToDecimal(ip)
	country, err := s.oracle.LookupCountry(ip)
	if err != nil {
		s.log.Debug(err)
	}
	countryISO, err := s.oracle.LookupCountryISO(ip)
	if err != nil {
		s.log.Debug(err)
	}
	city, err := s.oracle.LookupCity(ip)
	if err != nil {
		s.log.Debug(err)
	}
	hostnames, err := s.oracle.LookupAddr(ip)
	if err != nil {
		s.log.Debug(err)
	}
	return Response{
		IP:         ip,
		IPDecimal:  ipDecimal,
		Country:    country,
		CountryISO: countryISO,
		City:       city,
		Hostname:   strings.Join(hostnames, " "),
	}, nil
}

func (s *Server) newPortResponse(r *http.Request) (PortResponse, error) {
	vars := mux.Vars(r)
	port, err := strconv.ParseUint(vars["port"], 10, 16)
	if err != nil {
		return PortResponse{Port: port}, err
	}
	if port < 1 || port > 65355 {
		return PortResponse{Port: port}, fmt.Errorf("invalid port: %d", port)
	}
	ip, err := ipFromRequest(s.IPHeader, r)
	if err != nil {
		return PortResponse{Port: port}, err
	}
	err = s.oracle.LookupPort(ip, port)
	return PortResponse{
		IP:        ip,
		Port:      port,
		Reachable: err == nil,
	}, nil
}

func (s *Server) CLIHandler(w http.ResponseWriter, r *http.Request) *appError {
	ip, err := ipFromRequest(s.IPHeader, r)
	if err != nil {
		return internalServerError(err)
	}
	fmt.Fprintln(w, ip.String())
	return nil
}

func (s *Server) CLICountryHandler(w http.ResponseWriter, r *http.Request) *appError {
	response, err := s.newResponse(r)
	if err != nil {
		return internalServerError(err)
	}
	fmt.Fprintln(w, response.Country)
	return nil
}

func (s *Server) CLICountryISOHandler(w http.ResponseWriter, r *http.Request) *appError {
	response, err := s.newResponse(r)
	if err != nil {
		return internalServerError(err)
	}
	fmt.Fprintln(w, response.CountryISO)
	return nil
}

func (s *Server) CLICityHandler(w http.ResponseWriter, r *http.Request) *appError {
	response, err := s.newResponse(r)
	if err != nil {
		return internalServerError(err)
	}
	fmt.Fprintln(w, response.City)
	return nil
}

func (s *Server) JSONHandler(w http.ResponseWriter, r *http.Request) *appError {
	response, err := s.newResponse(r)
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

func (s *Server) PortHandler(w http.ResponseWriter, r *http.Request) *appError {
	response, err := s.newPortResponse(r)
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

func (s *Server) DefaultHandler(w http.ResponseWriter, r *http.Request) *appError {
	response, err := s.newResponse(r)
	if err != nil {
		return internalServerError(err)
	}
	t, err := template.New(filepath.Base(s.Template)).ParseFiles(s.Template)
	if err != nil {
		return internalServerError(err)
	}
	var data = struct {
		Host string
		Response
		Oracle
	}{r.Host, response, s.oracle}
	if err := t.Execute(w, &data); err != nil {
		return internalServerError(err)
	}
	return nil
}

func (s *Server) NotFoundHandler(w http.ResponseWriter, r *http.Request) *appError {
	err := notFound(nil).WithMessage("404 page not found")
	if r.Header.Get("accept") == jsonMediaType {
		err = err.AsJSON()
	}
	return err
}

func cliMatcher(r *http.Request, rm *mux.RouteMatch) bool {
	ua := useragent.Parse(r.UserAgent())
	switch ua.Product {
	case "curl", "HTTPie", "Wget", "fetch libfetch", "Go", "Go-http-client", "ddclient":
		return true
	}
	return false
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
		fmt.Fprint(w, e.Message)
	}
}

func (s *Server) Handler() http.Handler {
	r := mux.NewRouter()

	// JSON
	r.Handle("/", appHandler(s.JSONHandler)).Methods("GET").Headers("Accept", jsonMediaType)
	r.Handle("/json", appHandler(s.JSONHandler)).Methods("GET")

	// CLI
	r.Handle("/", appHandler(s.CLIHandler)).Methods("GET").MatcherFunc(cliMatcher)
	r.Handle("/", appHandler(s.CLIHandler)).Methods("GET").Headers("Accept", textMediaType)
	r.Handle("/ip", appHandler(s.CLIHandler)).Methods("GET")
	r.Handle("/country", appHandler(s.CLICountryHandler)).Methods("GET")
	r.Handle("/country-iso", appHandler(s.CLICountryISOHandler)).Methods("GET")
	r.Handle("/city", appHandler(s.CLICityHandler)).Methods("GET")

	// Browser
	r.Handle("/", appHandler(s.DefaultHandler)).Methods("GET")

	// Port testing
	r.Handle("/port/{port:[0-9]+}", appHandler(s.PortHandler)).Methods("GET")

	// Not found handler which returns JSON when appropriate
	r.NotFoundHandler = appHandler(s.NotFoundHandler)

	return r
}

func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s.Handler())
}
