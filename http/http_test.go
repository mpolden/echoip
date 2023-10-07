package http

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/levelsoftware/echoip/cache"
	"github.com/levelsoftware/echoip/iputil"
	"github.com/levelsoftware/echoip/iputil/geo"
	parser "github.com/levelsoftware/echoip/iputil/paser"
)

func lookupAddr(net.IP) (string, error) { return "localhost", nil }
func lookupPort(net.IP, uint64) error   { return nil }

type testDb struct{}

func (t *testDb) Country(net.IP) (geo.Country, error) {
	return geo.Country{Name: "Elbonia", ISO: "EB", IsEU: new(bool)}, nil
}

func (t *testDb) City(net.IP) (geo.City, error) {
	return geo.City{Name: "Bornyasherk", RegionName: "North Elbonia", RegionCode: "1234", MetroCode: 1234, PostalCode: "1234", Latitude: 63.416667, Longitude: 10.416667, Timezone: "Europe/Bornyasherk"}, nil
}

func (t *testDb) ASN(net.IP) (geo.ASN, error) {
	return geo.ASN{AutonomousSystemNumber: 59795, AutonomousSystemOrganization: "Hosting4Real"}, nil
}

func (t *testDb) IsEmpty() bool { return false }

func (t *testDb) Parse(ip net.IP, hostname string) (parser.Response, error) {
	ipDecimal := iputil.ToDecimal(ip)
	country, _ := t.Country(ip)
	city, _ := t.City(ip)
	asn, _ := t.ASN(ip)
	var autonomousSystemNumber string
	if asn.AutonomousSystemNumber > 0 {
		autonomousSystemNumber = fmt.Sprintf("AS%d", asn.AutonomousSystemNumber)
	}
	return parser.Response{
		UsingGeoIP:   true,
		UsingIPStack: false,
		IP:           ip,
		IPDecimal:    ipDecimal,
		Country:      country.Name,
		CountryISO:   country.ISO,
		CountryEU:    country.IsEU,
		RegionName:   city.RegionName,
		RegionCode:   city.RegionCode,
		MetroCode:    city.MetroCode,
		PostalCode:   city.PostalCode,
		City:         city.Name,
		Latitude:     city.Latitude,
		Longitude:    city.Longitude,
		Timezone:     city.Timezone,
		ASN:          autonomousSystemNumber,
		ASNOrg:       asn.AutonomousSystemOrganization,
		Hostname:     hostname,
	}, nil
}

type FakeCache struct{}

var cachedResponse cache.CachedResponse

func (fc *FakeCache) Get(ctx context.Context, ip string, response *cache.CachedResponse) error {
	response = &cachedResponse
	return nil
}

func (fc *FakeCache) Set(ctx context.Context, ip string, response cache.CachedResponse, cacheTtl int) error {
	cachedResponse = response
	return nil
}

func testServer() *Server {
	fakeCache := FakeCache{}
	return &Server{cache: &fakeCache, cacheTtl: 100, parser: &testDb{}, LookupAddr: lookupAddr, LookupPort: lookupPort}
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

func httpPost(url, body string) (*http.Response, string, error) {
	r, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		return nil, "", err
	}
	res, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, "", err
	}
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, "", err
	}
	return res, string(data), nil
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
		{s.URL + "/coordinates", "63.416667,10.416667\n", 200, "", ""},
		{s.URL + "/city", "Bornyasherk\n", 200, "", ""},
		{s.URL + "/foo", "404 page not found", 404, "", ""},
		{s.URL + "/asn", "AS59795\n", 200, "", ""},
		{s.URL + "/asn-org", "Hosting4Real\n", 200, "", ""},
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
	server.LookupPort = nil
	server.LookupAddr = nil
	parser, _ := geo.Open("", "", "")
	server.parser = &parser
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
		{s.URL + "/json", "{\n  \"UsingGeoIP\": true,\n  \"UsingIPStack\": false,\n  \"IPStackSecurityEnabled\": false,\n  \"timezone_etc\": {},\n  \"security\": {\n    \"is_proxy\": false,\n    \"is_crawler\": false,\n    \"is_tor\": false\n  },\n  \"currency\": {},\n  \"location\": {\n    \"country_flag\": {}\n  },\n  \"ip\": \"127.0.0.1\",\n  \"ip_decimal\": 2130706433\n}", 200},
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
		{s.URL, "{\n  \"UsingGeoIP\": true,\n  \"UsingIPStack\": false,\n  \"IPStackSecurityEnabled\": false,\n  \"timezone_etc\": {},\n  \"security\": {\n    \"is_proxy\": false,\n    \"is_crawler\": false,\n    \"is_tor\": false\n  },\n  \"currency\": {},\n  \"location\": {\n    \"country_flag\": {}\n  },\n  \"ip\": \"127.0.0.1\",\n  \"ip_decimal\": 2130706433,\n  \"country\": \"Elbonia\",\n  \"country_iso\": \"EB\",\n  \"country_eu\": false,\n  \"region_name\": \"North Elbonia\",\n  \"region_code\": \"1234\",\n  \"metro_code\": 1234,\n  \"zip_code\": \"1234\",\n  \"city\": \"Bornyasherk\",\n  \"latitude\": 63.416667,\n  \"longitude\": 10.416667,\n  \"timezone\": \"Europe/Bornyasherk\",\n  \"asn\": \"AS59795\",\n  \"asn_org\": \"Hosting4Real\",\n  \"hostname\": \"localhost\",\n  \"user_agent\": {\n    \"product\": \"curl\",\n    \"version\": \"7.2.6.0\",\n    \"raw_value\": \"curl/7.2.6.0\"\n  }\n}", 200},
		{s.URL + "/port/foo", "{\n  \"status\": 400,\n  \"error\": \"invalid port: foo\"\n}", 400},
		{s.URL + "/port/0", "{\n  \"status\": 400,\n  \"error\": \"invalid port: 0\"\n}", 400},
		{s.URL + "/port/65537", "{\n  \"status\": 400,\n  \"error\": \"invalid port: 65537\"\n}", 400},
		{s.URL + "/port/31337", "{\n  \"ip\": \"127.0.0.1\",\n  \"port\": 31337,\n  \"reachable\": true\n}", 200},
		{s.URL + "/port/80", "{\n  \"ip\": \"127.0.0.1\",\n  \"port\": 80,\n  \"reachable\": true\n}", 200},            // checking that our test server is reachable on port 80
		{s.URL + "/port/80?ip=1.3.3.7", "{\n  \"ip\": \"127.0.0.1\",\n  \"port\": 80,\n  \"reachable\": true\n}", 200}, // ensuring that the "ip" parameter is not usable to check remote host ports
		{s.URL + "/foo", "{\n  \"status\": 404,\n  \"error\": \"404 page not found\"\n}", 404},
		{s.URL + "/health", `{"status":"OK"}`, 200},
	}

	for _, tt := range tests {
		out, status, err := httpGet(tt.url, jsonMediaType, "curl/7.2.6.0")
		if err != nil {
			t.Fatal(err)
		}
		if status != tt.status {
			t.Errorf("Expected %d for %s, got %d", tt.status, tt.url, status)
		}
		if out != tt.out {
			t.Errorf("Expected %q for %s, got %q", tt.out, tt.url, out)
		}
	}
}

func TestIPFromRequest(t *testing.T) {
	var tests = []struct {
		remoteAddr     string
		headerKey      string
		headerValue    string
		trustedHeaders []string
		out            string
	}{
		{"127.0.0.1:9999", "", "", nil, "127.0.0.1"},                                                                // No header given
		{"127.0.0.1:9999", "X-Real-IP", "1.3.3.7", nil, "127.0.0.1"},                                                // Trusted header is empty
		{"127.0.0.1:9999", "X-Real-IP", "1.3.3.7", []string{"X-Foo-Bar"}, "127.0.0.1"},                              // Trusted header does not match
		{"127.0.0.1:9999", "X-Real-IP", "1.3.3.7", []string{"X-Real-IP", "X-Forwarded-For"}, "1.3.3.7"},             // Trusted header matches
		{"127.0.0.1:9999", "X-Forwarded-For", "1.3.3.7", []string{"X-Real-IP", "X-Forwarded-For"}, "1.3.3.7"},       // Second trusted header matches
		{"127.0.0.1:9999", "X-Forwarded-For", "1.3.3.7,4.2.4.2", []string{"X-Forwarded-For"}, "1.3.3.7"},            // X-Forwarded-For with multiple entries (commas separator)
		{"127.0.0.1:9999", "X-Forwarded-For", "1.3.3.7, 4.2.4.2", []string{"X-Forwarded-For"}, "1.3.3.7"},           // X-Forwarded-For with multiple entries (space+comma separator)
		{"127.0.0.1:9999", "X-Forwarded-For", "", []string{"X-Forwarded-For"}, "127.0.0.1"},                         // Empty header
		{"127.0.0.1:9999?ip=1.2.3.4", "", "", nil, "1.2.3.4"},                                                       // passed in "ip" parameter
		{"127.0.0.1:9999?ip=1.2.3.4", "X-Forwarded-For", "1.3.3.7,4.2.4.2", []string{"X-Forwarded-For"}, "1.2.3.4"}, // ip parameter wins over X-Forwarded-For with multiple entries
	}
	for _, tt := range tests {
		u, err := url.Parse("http://" + tt.remoteAddr)
		if err != nil {
			t.Fatal(err)
		}
		r := &http.Request{
			RemoteAddr: u.Host,
			Header:     http.Header{},
			URL:        u,
		}
		r.Header.Add(tt.headerKey, tt.headerValue)
		ip, err := ipFromRequest(tt.trustedHeaders, r, true)
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
		{"httpie-go/0.6.0", true},
		{"Go 1.1 package http", true},
		{"Go-http-client/1.1", true},
		{"Go-http-client/2.0", true},
		{"ddclient/3.8.3", true},
		{"Mikrotik/6.x Fetch", true},
		{browserUserAgent, false},
	}
	for _, tt := range tests {
		r := &http.Request{Header: http.Header{"User-Agent": []string{tt.in}}}
		if got := cliMatcher(r); got != tt.out {
			t.Errorf("Expected %t, got %t for %q", tt.out, got, tt.in)
		}
	}
}
