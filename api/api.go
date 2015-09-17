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
	IP_HEADER      = "x-ifconfig-ip"
	COUNTRY_HEADER = "x-ifconfig-country"
)

var cliUserAgentExp = regexp.MustCompile("^(?i)(curl|wget|fetch\\slibfetch)\\/.*$")

type API struct {
	db       *geoip2.Reader
	CORS     bool
	Template string
}

func New() *API { return &API{} }

func NewWithGeoIP(filepath string) (*API, error) {
	db, err := geoip2.Open(filepath)
	if err != nil {
		return nil, err
	}
	return &API{db: db}, nil
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

func headerKeyFromRequest(r *http.Request) string {
	vars := mux.Vars(r)
	key, ok := vars["key"]
	if !ok {
		return ""
	}
	return key
}

func (a *API) LookupCountry(ip net.IP) (string, error) {
	if a.db == nil {
		return "", nil
	}
	record, err := a.db.Country(ip)
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

func (a *API) defaultHandler(w http.ResponseWriter, r *http.Request) {
	cmd := cmdFromQueryParams(r.URL.Query())
	funcMap := template.FuncMap{"ToLower": strings.ToLower}
	t, err := template.New(filepath.Base(a.Template)).Funcs(funcMap).ParseFiles(a.Template)
	if err != nil {
		log.Print(err)
		return
	}
	b, err := json.MarshalIndent(r.Header, "", "  ")
	if err != nil {
		log.Print(err)
		return
	}

	var data = struct {
		IP     string
		JSON   string
		Header http.Header
		Cmd
	}{r.Header.Get(IP_HEADER), string(b), r.Header, cmd}

	if err := t.Execute(w, &data); err != nil {
		log.Print(err)
	}
}

func (a *API) jsonHandler(w http.ResponseWriter, r *http.Request) {
	key := headerKeyFromRequest(r)
	if key == "" {
		key = IP_HEADER
	}
	value := map[string]string{key: r.Header.Get(key)}
	b, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		log.Print(err)
		return
	}
	w.Write(b)
}

func (a *API) cliHandler(w http.ResponseWriter, r *http.Request) {
	key := headerKeyFromRequest(r)
	if key == "" {
		key = IP_HEADER
	}
	value := r.Header.Get(key)
	if !strings.HasSuffix(value, "\n") {
		value += "\n"
	}
	io.WriteString(w, value)
}

func cliMatcher(r *http.Request, rm *mux.RouteMatch) bool {
	return cliUserAgentExp.MatchString(r.UserAgent())
}

func (a *API) requestFilter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, err := ipFromRequest(r)
		if err != nil {
			r.Header.Set(IP_HEADER, err.Error())
		} else {
			r.Header.Set(IP_HEADER, ip.String())
			country, err := a.LookupCountry(ip)
			if err != nil {
				r.Header.Set(COUNTRY_HEADER, err.Error())
			} else {
				r.Header.Set(COUNTRY_HEADER, country)
			}
		}
		if a.CORS {
			w.Header().Set("Access-Control-Allow-Methods", "GET")
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		next.ServeHTTP(w, r)
	})
}

func (a *API) Handlers() http.Handler {
	r := mux.NewRouter()

	// JSON
	r.HandleFunc("/", a.jsonHandler).Methods("GET").Headers("Accept", "application/json")
	r.HandleFunc("/{key}", a.jsonHandler).Methods("GET").Headers("Accept", "application/json")
	r.HandleFunc("/{key}.json", a.jsonHandler).Methods("GET")

	// CLI
	r.HandleFunc("/", a.cliHandler).Methods("GET").MatcherFunc(cliMatcher)
	r.HandleFunc("/{key}", a.cliHandler).Methods("GET").MatcherFunc(cliMatcher)

	// Default
	r.HandleFunc("/", a.defaultHandler).Methods("GET")

	// Pass all requests through the request filter
	return a.requestFilter(r)
}

func (a *API) ListenAndServe(addr string) error {
	http.Handle("/", a.Handlers())
	return http.ListenAndServe(addr, nil)
}
