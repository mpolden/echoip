package http

import (
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mpolden/ipd/iputil/db"
)

func lookupAddr(net.IP) ([]string, error) { return []string{"localhost"}, nil }
func lookupPort(net.IP, uint64) error     { return nil }

type testDb struct{}

func (t *testDb) Country(net.IP) (db.Country, error) {
	return db.Country{Name: "Elbonia", ISO: "EB"}, nil
}

func (t *testDb) City(net.IP) (string, error) { return "Bornyasherk", nil }
func (t *testDb) IsEmpty() bool               { return false }

func testServer() *Server {
	return &Server{db: &testDb{}, lookupAddr: lookupAddr, lookupPort: lookupPort}
}

func httpGet(url string, acceptMediaType string, userAgent string) (string, int, error) {
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", 0, err
	}
	if acceptMediaType != "" {
		r.Header.Set("Accept", acceptMediaType)
	}
	r.Header.Set("User-Agent", userAgent)
	res, err := http.DefaultClient.Do(r)
	if err != nil {
		return "", 0, err
	}
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", 0, err
	}
	return string(data), res.StatusCode, nil
}

func TestCLIHandlers(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	s := httptest.NewServer(testServer().Handler())

	var tests = []struct {
		url             string
		out             string
		status          int
		userAgent       string
		acceptMediaType string
	}{
		{s.URL, "127.0.0.1\n", 200, "curl/7.43.0", ""},
		{s.URL, "127.0.0.1\n", 200, "foo/bar", textMediaType},
		{s.URL + "/ip", "127.0.0.1\n", 200, "", ""},
		{s.URL + "/country", "Elbonia\n", 200, "", ""},
		{s.URL + "/country-iso", "EB\n", 200, "", ""},
		{s.URL + "/city", "Bornyasherk\n", 200, "", ""},
		{s.URL + "/foo", "404 page not found", 404, "", ""},
	}

	for _, tt := range tests {
		out, status, err := httpGet(tt.url, tt.acceptMediaType, tt.userAgent)
		if err != nil {
			t.Fatal(err)
		}
		if status != tt.status {
			t.Errorf("Expected %d, got %d", tt.status, status)
		}
		if out != tt.out {
			t.Errorf("Expected %q, got %q", tt.out, out)
		}
	}
}

func TestDisabledHandlers(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	server := testServer()
	server.lookupPort = nil
	server.lookupAddr = nil
	server.db = db.Empty()
	s := httptest.NewServer(server.Handler())

	var tests = []struct {
		url    string
		out    string
		status int
	}{
		{s.URL + "/port/1337", "404 page not found", 404},
		{s.URL + "/country", "404 page not found", 404},
		{s.URL + "/country-iso", "404 page not found", 404},
		{s.URL + "/city", "404 page not found", 404},
		{s.URL + "/json", `{"ip":"127.0.0.1","ip_decimal":2130706433}`, 200},
	}

	for _, tt := range tests {
		out, status, err := httpGet(tt.url, "", "")
		if err != nil {
			t.Fatal(err)
		}
		if status != tt.status {
			t.Errorf("Expected %d, got %d", tt.status, status)
		}
		if out != tt.out {
			t.Errorf("Expected %q, got %q", tt.out, out)
		}
	}
}

func TestJSONHandlers(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	s := httptest.NewServer(testServer().Handler())

	var tests = []struct {
		url    string
		out    string
		status int
	}{
		{s.URL, `{"ip":"127.0.0.1","ip_decimal":2130706433,"country":"Elbonia","country_iso":"EB","city":"Bornyasherk","hostname":"localhost"}`, 200},
		{s.URL + "/port/foo", `{"error":"404 page not found"}`, 404},
		{s.URL + "/port/0", `{"error":"Invalid port: 0"}`, 400},
		{s.URL + "/port/65356", `{"error":"Invalid port: 65356"}`, 400},
		{s.URL + "/port/31337", `{"ip":"127.0.0.1","port":31337,"reachable":true}`, 200},
		{s.URL + "/foo", `{"error":"404 page not found"}`, 404},
	}

	for _, tt := range tests {
		out, status, err := httpGet(tt.url, jsonMediaType, "curl/7.2.6.0")
		if err != nil {
			t.Fatal(err)
		}
		if status != tt.status {
			t.Errorf("Expected %d, got %d", tt.status, status)
		}
		if out != tt.out {
			t.Errorf("Expected %q, got %q", tt.out, out)
		}
	}
}

func TestIPFromRequest(t *testing.T) {
	var tests = []struct {
		remoteAddr    string
		headerKey     string
		headerValue   string
		trustedHeader string
		out           string
	}{
		{"127.0.0.1:9999", "", "", "", "127.0.0.1"},                          // No header given
		{"127.0.0.1:9999", "X-Real-IP", "1.3.3.7", "", "127.0.0.1"},          // Trusted header is empty
		{"127.0.0.1:9999", "X-Real-IP", "1.3.3.7", "X-Foo-Bar", "127.0.0.1"}, // Trusted header does not match
		{"127.0.0.1:9999", "X-Real-IP", "1.3.3.7", "X-Real-IP", "1.3.3.7"},   // Trusted header matches
	}
	for _, tt := range tests {
		r := &http.Request{
			RemoteAddr: tt.remoteAddr,
			Header:     http.Header{},
		}
		r.Header.Add(tt.headerKey, tt.headerValue)
		ip, err := ipFromRequest(tt.trustedHeader, r)
		if err != nil {
			t.Fatal(err)
		}
		out := net.ParseIP(tt.out)
		if !ip.Equal(out) {
			t.Errorf("Expected %s, got %s", out, ip)
		}
	}
}

func TestCLIMatcher(t *testing.T) {
	browserUserAgent := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_8_4) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/30.0.1599.28 " +
		"Safari/537.36"
	var tests = []struct {
		in  string
		out bool
	}{
		{"curl/7.26.0", true},
		{"Wget/1.13.4 (linux-gnu)", true},
		{"Wget", true},
		{"fetch libfetch/2.0", true},
		{"HTTPie/0.9.3", true},
		{"Go 1.1 package http", true},
		{"Go-http-client/1.1", true},
		{"Go-http-client/2.0", true},
		{"ddclient/3.8.3", true},
		{browserUserAgent, false},
	}
	for _, tt := range tests {
		r := &http.Request{Header: http.Header{"User-Agent": []string{tt.in}}}
		if got := cliMatcher(r, nil); got != tt.out {
			t.Errorf("Expected %t, got %t for %q", tt.out, got, tt.in)
		}
	}
}
