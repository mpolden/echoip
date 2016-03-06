package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	"html/template"

	"github.com/gorilla/mux"
	geoip2 "github.com/oschwald/geoip2-golang"
)

const (
	IP_HEADER        = "x-ifconfig-ip"
	COUNTRY_HEADER   = "x-ifconfig-country"
	HOSTNAME_HEADER  = "x-ifconfig-hostname"
	TEXT_PLAIN       = "text/plain; charset=utf-8"
	APPLICATION_JSON = "application/json"
)

var cliUserAgentExp = regexp.MustCompile(`^(?i)((curl|wget|fetch\slibfetch|Go-http-client)\/.*|Go\s1\.1\spackage\shttp)$`)

type API struct {
	CORS          bool
	ReverseLookup bool
	Template      string
	lookupAddr    func(string) ([]string, error)
	lookupCountry func(net.IP) (string, error)
	ipFromRequest func(*http.Request) (net.IP, error)
}

func New() *API {
	return &API{
		lookupAddr:    net.LookupAddr,
		lookupCountry: func(ip net.IP) (string, error) { return "", nil },
		ipFromRequest: ipFromRequest,
	}
}

func NewWithGeoIP(filepath string) (*API, error) {
	db, err := geoip2.Open(filepath)
	if err != nil {
		return nil, err
	}
	api := New()
	api.lookupCountry = func(ip net.IP) (string, error) {
		return lookupCountry(db, ip)
	}
	return api, nil
}

type Cmd struct {
	Name string
	Args string
}

func (c *Cmd) String() string {
	return c.Name + " " + c.Args
}

func cmdFromQueryParams(values url.Values) Cmd {
	cmd, exists := values["cmd"]
	if !exists || len(cmd) == 0 {
		return Cmd{Name: "curl"}
	}
	switch cmd[0] {
	case "fetch":
		return Cmd{Name: "fetch", Args: "-qo -"}
	case "wget":
		return Cmd{Name: "wget", Args: "-qO -"}
	}
	return Cmd{Name: "curl"}
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

func headerPairFromRequest(r *http.Request) (string, string, error) {
	vars := mux.Vars(r)
	header, ok := vars["header"]
	if !ok {
		header = IP_HEADER
	}
	value := r.Header.Get(header)
	if value == "" {
		return "", "", fmt.Errorf("no value found for: %s", header)
	}
	return header, value, nil
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
	return "", fmt.Errorf("could not determine country for IP: %s", ip)
}

func (a *API) DefaultHandler(w http.ResponseWriter, r *http.Request) *appError {
	cmd := cmdFromQueryParams(r.URL.Query())
	funcMap := template.FuncMap{"ToLower": strings.ToLower}
	t, err := template.New(filepath.Base(a.Template)).Funcs(funcMap).ParseFiles(a.Template)
	if err != nil {
		return internalServerError(err)
	}
	b, err := json.MarshalIndent(r.Header, "", "  ")
	if err != nil {
		return internalServerError(err)
	}

	IsV6 := true
	ip := net.ParseIP(r.Header.Get(IP_HEADER))
	if ip.To4() != nil {
		IsV6 = false
	}

	var data = struct {
		IP     string
		IsV6   bool
		JSON   string
		Header http.Header
		Cmd
	}{ip.String(), IsV6, string(b), r.Header, cmd}

	if err := t.Execute(w, &data); err != nil {
		return internalServerError(err)
	}
	return nil
}

func (a *API) JSONHandler(w http.ResponseWriter, r *http.Request) *appError {
	k, v, err := headerPairFromRequest(r)
	contentType := APPLICATION_JSON
	if err != nil {
		return notFound(err).WithContentType(contentType)
	}
	value := map[string]string{k: v}
	b, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return internalServerError(err).WithContentType(contentType)
	}
	w.Header().Set("Content-Type", contentType)
	w.Write(b)
	return nil
}

func (a *API) JSONAllHandler(w http.ResponseWriter, r *http.Request) *appError {
	contentType := APPLICATION_JSON
	b, err := json.MarshalIndent(r.Header, "", "  ")
	if err != nil {
		return internalServerError(err).WithContentType(contentType)
	}
	w.Header().Set("Content-Type", contentType)
	w.Write(b)
	return nil
}

func (a *API) CLIHandler(w http.ResponseWriter, r *http.Request) *appError {
	_, v, err := headerPairFromRequest(r)
	if err != nil {
		return notFound(err)
	}
	if !strings.HasSuffix(v, "\n") {
		v += "\n"
	}
	io.WriteString(w, v)
	return nil
}

func cliMatcher(r *http.Request, rm *mux.RouteMatch) bool {
	return cliUserAgentExp.MatchString(r.UserAgent())
}

func (a *API) requestFilter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, err := a.ipFromRequest(r)
		if err != nil {
			log.Print(err)
			r.Header.Set(IP_HEADER, "")
		} else {
			r.Header.Set(IP_HEADER, ip.String())
			country, err := a.lookupCountry(ip)
			if err != nil {
				log.Print(err)
			}
			r.Header.Set(COUNTRY_HEADER, country)
		}
		if a.ReverseLookup {
			hostname, err := a.lookupAddr(ip.String())
			if err != nil {
				log.Print(err)
			}
			r.Header.Set(HOSTNAME_HEADER, strings.Join(hostname, ", "))
		}
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
		contentType := e.ContentType
		if contentType == "" {
			contentType = TEXT_PLAIN
		}
		response := e.Response
		if response == "" {
			response = e.Error.Error()
		}
		if e.IsJSON() {
			var data = struct {
				Error string `json:"error"`
			}{response}
			b, err := json.MarshalIndent(data, "", "  ")
			if err != nil {
				panic(err)
			}
			response = string(b)
		}
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(e.Code)
		io.WriteString(w, response)
	}
}

func (a *API) Handlers() http.Handler {
	r := mux.NewRouter()

	// JSON
	r.Handle("/", appHandler(a.JSONHandler)).Methods("GET").Headers("Accept", APPLICATION_JSON)
	r.Handle("/all", appHandler(a.JSONAllHandler)).Methods("GET").Headers("Accept", APPLICATION_JSON)
	r.Handle("/all.json", appHandler(a.JSONAllHandler)).Methods("GET")
	r.Handle("/{header}", appHandler(a.JSONHandler)).Methods("GET").Headers("Accept", APPLICATION_JSON)
	r.Handle("/{header}.json", appHandler(a.JSONHandler)).Methods("GET")

	// CLI
	r.Handle("/", appHandler(a.CLIHandler)).Methods("GET").MatcherFunc(cliMatcher)
	r.Handle("/{header}", appHandler(a.CLIHandler)).Methods("GET").MatcherFunc(cliMatcher)

	// Default
	r.Handle("/", appHandler(a.DefaultHandler)).Methods("GET")

	// Pass all requests through the request filter
	return a.requestFilter(r)
}

func (a *API) ListenAndServe(addr string) error {
	http.Handle("/", a.Handlers())
	return http.ListenAndServe(addr, nil)
}
