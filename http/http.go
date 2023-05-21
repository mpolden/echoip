package http

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"net/http/pprof"

	"github.com/mpolden/echoip/iputil"
	"github.com/mpolden/echoip/iputil/geo"
	"github.com/mpolden/echoip/useragent"

	"math/big"
	"net"
	"net/http"
	"strconv"
)

const (
	jsonMediaType = "application/json"
	textMediaType = "text/plain"
)

type Server struct {
	Template   string
	IPHeaders  []string
	LookupAddr func(net.IP) (string, error)
	LookupPort func(net.IP, uint64) error
	cache      *Cache
	gr         geo.Reader
	profile    bool
	Sponsor    bool
}

type Response struct {
	IP         net.IP               `json:"ip"`
	IPDecimal  *big.Int             `json:"ip_decimal"`
	Country    string               `json:"country,omitempty"`
	CountryISO string               `json:"country_iso,omitempty"`
	CountryEU  *bool                `json:"country_eu,omitempty"`
	RegionName string               `json:"region_name,omitempty"`
	RegionCode string               `json:"region_code,omitempty"`
	MetroCode  uint                 `json:"metro_code,omitempty"`
	PostalCode string               `json:"zip_code,omitempty"`
	City       string               `json:"city,omitempty"`
	Latitude   float64              `json:"latitude,omitempty"`
	Longitude  float64              `json:"longitude,omitempty"`
	Timezone   string               `json:"time_zone,omitempty"`
	ASN        string               `json:"asn,omitempty"`
	ASNOrg     string               `json:"asn_org,omitempty"`
	Hostname   string               `json:"hostname,omitempty"`
	UserAgent  *useragent.UserAgent `json:"user_agent,omitempty"`
}

type PortResponse struct {
	IP        net.IP `json:"ip"`
	Port      uint64 `json:"port"`
	Reachable bool   `json:"reachable"`
}

func New(db geo.Reader, cache *Cache, profile bool) *Server {
	return &Server{cache: cache, gr: db, profile: profile}
}

func ipFromForwardedForHeader(v string) string {
	sep := strings.Index(v, ",")
	if sep == -1 {
		return v
	}
	return v[:sep]
}

// ipFromRequest detects the IP address for this transaction.
//
// * `headers` - the specific HTTP headers to trust
// * `r` - the incoming HTTP request
// * `customIP` - whether to allow the IP to be pulled from query parameters
func ipFromRequest(headers []string, r *http.Request, customIP bool) (net.IP, error) {
	remoteIP := ""
	if customIP && r.URL != nil {
		if v, ok := r.URL.Query()["ip"]; ok {
			remoteIP = v[0]
		}
	}
	if remoteIP == "" {
		for _, header := range headers {
			remoteIP = r.Header.Get(header)
			if http.CanonicalHeaderKey(header) == "X-Forwarded-For" {
				remoteIP = ipFromForwardedForHeader(remoteIP)
			}
			if remoteIP != "" {
				break
			}
		}
	}
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

func userAgentFromRequest(r *http.Request) *useragent.UserAgent {
	var userAgent *useragent.UserAgent
	userAgentRaw := r.UserAgent()
	if userAgentRaw != "" {
		parsed := useragent.Parse(userAgentRaw)
		userAgent = &parsed
	}
	return userAgent
}

func (s *Server) newResponse(r *http.Request) (Response, error) {
	ip, err := ipFromRequest(s.IPHeaders, r, true)
	if err != nil {
		return Response{}, err
	}
	response, ok := s.cache.Get(ip)
	if ok {
		// Do not cache user agent
		response.UserAgent = userAgentFromRequest(r)
		return response, nil
	}
	ipDecimal := iputil.ToDecimal(ip)
	country, _ := s.gr.Country(ip)
	city, _ := s.gr.City(ip)
	asn, _ := s.gr.ASN(ip)
	var hostname string
	if s.LookupAddr != nil {
		hostname, _ = s.LookupAddr(ip)
	}
	var autonomousSystemNumber string
	if asn.AutonomousSystemNumber > 0 {
		autonomousSystemNumber = fmt.Sprintf("AS%d", asn.AutonomousSystemNumber)
	}
	response = Response{
		IP:         ip,
		IPDecimal:  ipDecimal,
		Country:    country.Name,
		CountryISO: country.ISO,
		CountryEU:  country.IsEU,
		RegionName: city.RegionName,
		RegionCode: city.RegionCode,
		MetroCode:  city.MetroCode,
		PostalCode: city.PostalCode,
		City:       city.Name,
		Latitude:   city.Latitude,
		Longitude:  city.Longitude,
		Timezone:   city.Timezone,
		ASN:        autonomousSystemNumber,
		ASNOrg:     asn.AutonomousSystemOrganization,
		Hostname:   hostname,
	}
	s.cache.Set(ip, response)
	response.UserAgent = userAgentFromRequest(r)
	return response, nil
}

func (s *Server) newPortResponse(r *http.Request) (PortResponse, error) {
	lastElement := filepath.Base(r.URL.Path)
	port, err := strconv.ParseUint(lastElement, 10, 16)
	if err != nil || port < 1 || port > 65535 {
		return PortResponse{Port: port}, fmt.Errorf("invalid port: %s", lastElement)
	}
	ip, err := ipFromRequest(s.IPHeaders, r, false)
	if err != nil {
		return PortResponse{Port: port}, err
	}
	err = s.LookupPort(ip, port)
	return PortResponse{
		IP:        ip,
		Port:      port,
		Reachable: err == nil,
	}, nil
}

func (s *Server) CLIHandler(w http.ResponseWriter, r *http.Request) *appError {
	ip, err := ipFromRequest(s.IPHeaders, r, true)
	if err != nil {
		return badRequest(err).WithMessage(err.Error()).AsJSON()
	}
	fmt.Fprintln(w, ip.String())
	return nil
}

func (s *Server) CLICountryHandler(w http.ResponseWriter, r *http.Request) *appError {
	response, err := s.newResponse(r)
	if err != nil {
		return badRequest(err).WithMessage(err.Error()).AsJSON()
	}
	fmt.Fprintln(w, response.Country)
	return nil
}

func (s *Server) CLICountryISOHandler(w http.ResponseWriter, r *http.Request) *appError {
	response, err := s.newResponse(r)
	if err != nil {
		return badRequest(err).WithMessage(err.Error()).AsJSON()
	}
	fmt.Fprintln(w, response.CountryISO)
	return nil
}

func (s *Server) CLICityHandler(w http.ResponseWriter, r *http.Request) *appError {
	response, err := s.newResponse(r)
	if err != nil {
		return badRequest(err).WithMessage(err.Error()).AsJSON()
	}
	fmt.Fprintln(w, response.City)
	return nil
}

func (s *Server) CLICoordinatesHandler(w http.ResponseWriter, r *http.Request) *appError {
	response, err := s.newResponse(r)
	if err != nil {
		return badRequest(err).WithMessage(err.Error()).AsJSON()
	}
	fmt.Fprintf(w, "%s,%s\n", formatCoordinate(response.Latitude), formatCoordinate(response.Longitude))
	return nil
}

func (s *Server) CLIASNHandler(w http.ResponseWriter, r *http.Request) *appError {
	response, err := s.newResponse(r)
	if err != nil {
		return badRequest(err).WithMessage(err.Error()).AsJSON()
	}
	fmt.Fprintf(w, "%s\n", response.ASN)
	return nil
}

func (s *Server) CLIASNOrgHandler(w http.ResponseWriter, r *http.Request) *appError {
	response, err := s.newResponse(r)
	if err != nil {
		return badRequest(err).WithMessage(err.Error()).AsJSON()
	}
	fmt.Fprintf(w, "%s\n", response.ASNOrg)
	return nil
}

func (s *Server) JSONHandler(w http.ResponseWriter, r *http.Request) *appError {
	response, err := s.newResponse(r)
	if err != nil {
		return badRequest(err).WithMessage(err.Error()).AsJSON()
	}
	b, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return internalServerError(err).AsJSON()
	}
	w.Header().Set("Content-Type", jsonMediaType)
	w.Write(b)
	return nil
}

func (s *Server) HealthHandler(w http.ResponseWriter, r *http.Request) *appError {
	w.Header().Set("Content-Type", jsonMediaType)
	w.Write([]byte(`{"status":"OK"}`))
	return nil
}

func (s *Server) PortHandler(w http.ResponseWriter, r *http.Request) *appError {
	response, err := s.newPortResponse(r)
	if err != nil {
		return badRequest(err).WithMessage(err.Error()).AsJSON()
	}
	b, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return internalServerError(err).AsJSON()
	}
	w.Header().Set("Content-Type", jsonMediaType)
	w.Write(b)
	return nil
}

func (s *Server) cacheResizeHandler(w http.ResponseWriter, r *http.Request) *appError {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return badRequest(err).WithMessage(err.Error()).AsJSON()
	}
	capacity, err := strconv.Atoi(string(body))
	if err != nil {
		return badRequest(err).WithMessage(err.Error()).AsJSON()
	}
	if err := s.cache.Resize(capacity); err != nil {
		return badRequest(err).WithMessage(err.Error()).AsJSON()
	}
	data := struct {
		Message string `json:"message"`
	}{fmt.Sprintf("Changed cache capacity to %d.", capacity)}
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return internalServerError(err).AsJSON()
	}
	w.Header().Set("Content-Type", jsonMediaType)
	w.Write(b)
	return nil
}

func (s *Server) cacheHandler(w http.ResponseWriter, r *http.Request) *appError {
	cacheStats := s.cache.Stats()
	var data = struct {
		Size      int    `json:"size"`
		Capacity  int    `json:"capacity"`
		Evictions uint64 `json:"evictions"`
	}{
		cacheStats.Size,
		cacheStats.Capacity,
		cacheStats.Evictions,
	}
	b, err := json.MarshalIndent(data, "", "  ")
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
		return badRequest(err).WithMessage(err.Error())
	}
	t, err := template.ParseGlob(s.Template + "/*")
	if err != nil {
		return internalServerError(err)
	}
	json, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return internalServerError(err)
	}

	var data = struct {
		Response
		Host         string
		BoxLatTop    float64
		BoxLatBottom float64
		BoxLonLeft   float64
		BoxLonRight  float64
		JSON         string
		Port         bool
		Sponsor      bool
	}{
		response,
		r.Host,
		response.Latitude + 0.05,
		response.Latitude - 0.05,
		response.Longitude - 0.05,
		response.Longitude + 0.05,
		string(json),
		s.LookupPort != nil,
		s.Sponsor,
	}
	if err := t.Execute(w, &data); err != nil {
		return internalServerError(err)
	}
	return nil
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) *appError {
	err := notFound(nil).WithMessage("404 page not found")
	if r.Header.Get("accept") == jsonMediaType {
		err = err.AsJSON()
	}
	return err
}

func cliMatcher(r *http.Request) bool {
	ua := useragent.Parse(r.UserAgent())
	switch ua.Product {
	case "curl", "HTTPie", "httpie-go", "Wget", "fetch libfetch", "Go", "Go-http-client", "ddclient", "Mikrotik", "xh":
		return true
	}
	return false
}

type appHandler func(http.ResponseWriter, *http.Request) *appError

func wrapHandlerFunc(f http.HandlerFunc) appHandler {
	return func(w http.ResponseWriter, r *http.Request) *appError {
		f.ServeHTTP(w, r)
		return nil
	}
}

func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if e := fn(w, r); e != nil { // e is *appError
		if e.Code/100 == 5 {
			log.Println(e.Error)
		}
		// When Content-Type for error is JSON, we need to marshal the response into JSON
		if e.IsJSON() {
			var data = struct {
				Code  int    `json:"status"`
				Error string `json:"error"`
			}{e.Code, e.Message}
			b, err := json.MarshalIndent(data, "", "  ")
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
	r := NewRouter()

	// Health
	r.Route("GET", "/health", s.HealthHandler)

	// JSON
	r.Route("GET", "/", s.JSONHandler).Header("Accept", jsonMediaType)
	r.Route("GET", "/json", s.JSONHandler)

	// CLI
	r.Route("GET", "/", s.CLIHandler).MatcherFunc(cliMatcher)
	r.Route("GET", "/", s.CLIHandler).Header("Accept", textMediaType)
	r.Route("GET", "/ip", s.CLIHandler)
	if !s.gr.IsEmpty() {
		r.Route("GET", "/country", s.CLICountryHandler)
		r.Route("GET", "/country-iso", s.CLICountryISOHandler)
		r.Route("GET", "/city", s.CLICityHandler)
		r.Route("GET", "/coordinates", s.CLICoordinatesHandler)
		r.Route("GET", "/asn", s.CLIASNHandler)
		r.Route("GET", "/asn-org", s.CLIASNOrgHandler)
	}

	// Browser
	if s.Template != "" {
		r.Route("GET", "/", s.DefaultHandler)
	}

	// Port testing
	if s.LookupPort != nil {
		r.RoutePrefix("GET", "/port/", s.PortHandler)
	}

	// Profiling
	if s.profile {
		r.Route("POST", "/debug/cache/resize", s.cacheResizeHandler)
		r.Route("GET", "/debug/cache/", s.cacheHandler)
		r.Route("GET", "/debug/pprof/cmdline", wrapHandlerFunc(pprof.Cmdline))
		r.Route("GET", "/debug/pprof/profile", wrapHandlerFunc(pprof.Profile))
		r.Route("GET", "/debug/pprof/symbol", wrapHandlerFunc(pprof.Symbol))
		r.Route("GET", "/debug/pprof/trace", wrapHandlerFunc(pprof.Trace))
		r.RoutePrefix("GET", "/debug/pprof/", wrapHandlerFunc(pprof.Index))
	}

	return r.Handler()
}

func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s.Handler())
}

func formatCoordinate(c float64) string {
	return strconv.FormatFloat(c, 'f', 6, 64)
}
