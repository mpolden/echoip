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
	"time"

	"github.com/gorilla/mux"
	geoip2 "github.com/oschwald/geoip2-golang"
)

const APPLICATION_JSON = "application/json"

var cliUserAgentExp = regexp.MustCompile(`^((curl|Wget|fetch\slibfetch|Go-http-client|HTTPie)\/.*|Go\s1\.1\spackage\shttp)$`)

type API struct {
	CORS          bool
	Template      string
	lookupAddr    func(string) ([]string, error)
	lookupCountry func(net.IP) (string, error)
	testPort      func(net.IP, uint64) error
	ipFromRequest func(*http.Request) (net.IP, error)
	reverseLookup bool
	countryLookup bool
	portTesting   bool
}

type Response struct {
	IP       net.IP `json:"ip"`
	Country  string `json:"country,omitempty"`
	Hostname string `json:"hostname,omitempty"`
}

type TestPortResponse struct {
	IP        net.IP `json:"ip"`
	Port      uint64 `json:"port"`
	Reachable bool   `json:"reachable"`
}

func New() *API {
	return &API{
		lookupAddr:    func(addr string) (names []string, err error) { return nil, nil },
		lookupCountry: func(ip net.IP) (string, error) { return "", nil },
		testPort:      func(ip net.IP, port uint64) error { return nil },
		ipFromRequest: ipFromRequest,
	}
}

func (a *API) EnableCountryLookup(filepath string) error {
	db, err := geoip2.Open(filepath)
	if err != nil {
		return err
	}
	a.lookupCountry = func(ip net.IP) (string, error) {
		return lookupCountry(db, ip)
	}
	a.countryLookup = true
	return nil
}

func (a *API) EnableReverseLookup() {
	a.lookupAddr = net.LookupAddr
	a.reverseLookup = true
}

func (a *API) EnablePortTesting() {
	a.testPort = testPort
	a.portTesting = true
}

func ipFromRequest(r *http.Request) (net.IP, error) {
	var host string
	realIP := r.Header.Get("X-Real-IP")
	var err error
	if realIP != "" {
		host = realIP
	} else {
		host, _, err = net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			return nil, err
		}
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return nil, fmt.Errorf("could not parse IP: %s", host)
	}
	return ip, nil
}

func testPort(ip net.IP, port uint64) error {
	address := fmt.Sprintf("%s:%d", ip, port)
	if _, err := net.DialTimeout("tcp", address, 2*time.Second); err != nil {
		return err
	}
	return nil
}

func lookupCountry(db *geoip2.Reader, ip net.IP) (string, error) {
	if db == nil {
		return "", nil
	}
	record, err := db.Country(ip)
	if err != nil {
		return "", err
	}
	if country, exists := record.Country.Names["en"]; exists {
		return country, nil
	}
	if country, exists := record.RegisteredCountry.Names["en"]; exists {
		return country, nil
	}
	return "Unknown", fmt.Errorf("could not determine country for IP: %s", ip)
}

func (a *API) newResponse(r *http.Request) (Response, error) {
	ip, err := a.ipFromRequest(r)
	if err != nil {
		return Response{}, err
	}
	country, err := a.lookupCountry(ip)
	if err != nil {
		log.Print(err)
	}
	hostnames, err := a.lookupAddr(ip.String())
	if err != nil {
		log.Print(err)
	}
	return Response{
		IP:       ip,
		Country:  country,
		Hostname: strings.Join(hostnames, " "),
	}, nil
}

func (a *API) CLIHandler(w http.ResponseWriter, r *http.Request) *appError {
	ip, err := a.ipFromRequest(r)
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

func (a *API) TestPortHandler(w http.ResponseWriter, r *http.Request) *appError {
	vars := mux.Vars(r)
	port, err := strconv.ParseUint(vars["port"], 10, 16)
	if err != nil {
		return badRequest(err).WithMessage("Invalid port: " + vars["port"]).AsJSON()
	}
	if port < 1 || port > 65355 {
		return badRequest(nil).WithMessage("Invalid port: " + vars["port"]).AsJSON()
	}
	ip, err := a.ipFromRequest(r)
	if err != nil {
		return internalServerError(err).AsJSON()
	}
	err = testPort(ip, port)
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
		ReverseLookup bool
		CountryLookup bool
		PortTesting   bool
	}{response, a.reverseLookup, a.countryLookup, a.portTesting}
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
	return cliUserAgentExp.MatchString(r.UserAgent())
}

func (a *API) requestFilter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.CORS {
			w.Header().Set("Access-Control-Allow-Methods", "GET")
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		next.ServeHTTP(w, r)
	})
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

	// CLI
	r.Handle("/", appHandler(a.CLIHandler)).Methods("GET").MatcherFunc(cliMatcher)
	r.Handle("/ip", appHandler(a.CLIHandler)).Methods("GET").MatcherFunc(cliMatcher)
	r.Handle("/country", appHandler(a.CLICountryHandler)).Methods("GET").MatcherFunc(cliMatcher)

	// Browser
	r.Handle("/", appHandler(a.DefaultHandler)).Methods("GET")

	// Port testing
	r.Handle("/port/{port:[0-9]+}", appHandler(a.TestPortHandler)).Methods("GET")

	// Not found handler which returns JSON when appropriate
	r.NotFoundHandler = appHandler(a.NotFoundHandler)

	// Pass all requests through the request filter
	return a.requestFilter(r)
}

func (a *API) ListenAndServe(addr string) error {
	http.Handle("/", a.Handlers())
	return http.ListenAndServe(addr, nil)
}
