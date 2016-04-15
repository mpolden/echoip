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
	"strings"

	"github.com/gorilla/mux"
	geoip2 "github.com/oschwald/geoip2-golang"
)

const APPLICATION_JSON = "application/json"

var cliUserAgentExp = regexp.MustCompile(`^(?i)((curl|wget|fetch\slibfetch|Go-http-client)\/.*|Go\s1\.1\spackage\shttp)$`)

type API struct {
	CORS          bool
	Template      string
	lookupAddr    func(string) ([]string, error)
	lookupCountry func(net.IP) (string, error)
	ipFromRequest func(*http.Request) (net.IP, error)
}

type Response struct {
	IP       net.IP `json:"ip"`
	Country  string `json:"country,omitempty"`
	Hostname string `json:"hostname,omitempty"`
}

func New() *API {
	return &API{
		lookupAddr:    func(addr string) (names []string, err error) { return nil, nil },
		lookupCountry: func(ip net.IP) (string, error) { return "", nil },
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
	return nil
}

func (a *API) EnableReverseLookup() {
	a.lookupAddr = net.LookupAddr
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
	response, err := a.newResponse(r)
	if err != nil {
		return internalServerError(err)
	}
	if r.URL.Path == "/country" {
		io.WriteString(w, response.Country+"\n")
	} else {
		io.WriteString(w, response.IP.String()+"\n")
	}
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

func (a *API) DefaultHandler(w http.ResponseWriter, r *http.Request) *appError {
	response, err := a.newResponse(r)
	if err != nil {
		return internalServerError(err)
	}
	t, err := template.New(filepath.Base(a.Template)).ParseFiles(a.Template)
	if err != nil {
		return internalServerError(err)
	}
	if err := t.Execute(w, &response); err != nil {
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
	r.Handle("/country", appHandler(a.CLIHandler)).Methods("GET").MatcherFunc(cliMatcher)

	// Browser
	r.Handle("/", appHandler(a.DefaultHandler)).Methods("GET")

	// Not found handler which returns JSON when appropriate
	r.NotFoundHandler = appHandler(a.NotFoundHandler)

	// Pass all requests through the request filter
	return a.requestFilter(r)
}

func (a *API) ListenAndServe(addr string) error {
	http.Handle("/", a.Handlers())
	return http.ListenAndServe(addr, nil)
}
