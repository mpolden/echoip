package http

import (
	"encoding/json"
	"fmt"
	"html/template"
	"path/filepath"
	"strings"

	"github.com/mpolden/ipd/iputil"
	"github.com/mpolden/ipd/iputil/geo"
	"github.com/mpolden/ipd/useragent"

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
	gr         geo.Reader
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

func New(db geo.Reader) *Server {
	return &Server{gr: db}
}

func ipFromForwardedForHeader(v string) string {
	sep := strings.Index(v, ",")
	if sep == -1 {
		return v
	}
	return v[:sep]
}

func ipFromRequest(headers []string, r *http.Request) (net.IP, error) {
	remoteIP := ""
	for _, header := range headers {
		remoteIP = r.Header.Get(header)
		if http.CanonicalHeaderKey(header) == "X-Forwarded-For" {
			remoteIP = ipFromForwardedForHeader(remoteIP)
		}
		if remoteIP != "" {
			break
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

func (s *Server) newResponse(r *http.Request) (Response, error) {
	ip, err := ipFromRequest(s.IPHeaders, r)
	if err != nil {
		return Response{}, err
	}
	ipDecimal := iputil.ToDecimal(ip)
	country, _ := s.gr.Country(ip)
	city, _ := s.gr.City(ip)
	var hostname string
	if s.LookupAddr != nil {
		hostname, _ = s.LookupAddr(ip)
	}
	return Response{
		IP:         ip,
		IPDecimal:  ipDecimal,
		Country:    country.Name,
		CountryISO: country.ISO,
		City:       city,
		Hostname:   hostname,
	}, nil
}

func (s *Server) newPortResponse(r *http.Request) (PortResponse, error) {
	lastElement := filepath.Base(r.URL.Path)
	port, err := strconv.ParseUint(lastElement, 10, 16)
	if err != nil || port < 1 || port > 65355 {
		return PortResponse{Port: port}, fmt.Errorf("invalid port: %d", port)
	}
	ip, err := ipFromRequest(s.IPHeaders, r)
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
	ip, err := ipFromRequest(s.IPHeaders, r)
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

func (s *Server) HealthHandler(w http.ResponseWriter, r *http.Request) *appError {
	w.Header().Set("Content-Type", jsonMediaType)
	w.Write([]byte(`{"status":"OK"}`))
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
	t, err := template.ParseFiles(s.Template)
	if err != nil {
		return internalServerError(err)
	}
	json, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return internalServerError(err)
	}
	var data = struct {
		Response
		Host string
		JSON string
		Port bool
	}{
		response,
		r.Host,
		string(json),
		s.LookupPort != nil,
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
	}

	// Browser
	r.Route("GET", "/", s.DefaultHandler)

	// Port testing
	if s.LookupPort != nil {
		r.RoutePrefix("GET", "/port/", s.PortHandler)
	}

	return r.Handler()
}

func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s.Handler())
}
