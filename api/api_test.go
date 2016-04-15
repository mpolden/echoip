package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestAPI() *API {
	return &API{
		lookupAddr: func(string) ([]string, error) {
			return []string{"localhost"}, nil
		},
		lookupCountry: func(ip net.IP) (string, error) {
			return "Elbonia", nil
		},
		ipFromRequest: func(*http.Request) (net.IP, error) {
			return net.ParseIP("127.0.0.1"), nil
		},
	}
}

func httpGet(url string, json bool, userAgent string) (string, int, error) {
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", 0, err
	}
	if json {
		r.Header.Set("Accept", "application/json")
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

func TestGetIP(t *testing.T) {
	//log.SetOutput(ioutil.Discard)
	toJSON := func(r Response) string {
		b, err := json.Marshal(r)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	s := httptest.NewServer(newTestAPI().Handlers())
	var tests = []struct {
		url       string
		json      bool
		out       string
		userAgent string
		status    int
	}{
		{s.URL, false, "127.0.0.1\n", "curl/7.26.0", 200},
		{s.URL, false, "127.0.0.1\n", "Wget/1.13.4 (linux-gnu)", 200},
		{s.URL, false, "127.0.0.1\n", "fetch libfetch/2.0", 200},
		{s.URL, false, "127.0.0.1\n", "Go 1.1 package http", 200},
		{s.URL, false, "127.0.0.1\n", "Go-http-client/1.1", 200},
		{s.URL, false, "127.0.0.1\n", "Go-http-client/2.0", 200},
		{s.URL, true, toJSON(Response{IP: net.ParseIP("127.0.0.1"), Country: "Elbonia", Hostname: "localhost"}), "", 200},
		{s.URL + "/foo", false, "404 page not found", "curl/7.26.0", 404},
		{s.URL + "/foo", true, "{\"error\":\"404 page not found\"}", "curl/7.26.0", 404},
	}

	for _, tt := range tests {
		out, status, err := httpGet(tt.url, tt.json, tt.userAgent)
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

func TestGetIPWithoutReverse(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	api := newTestAPI()
	s := httptest.NewServer(api.Handlers())

	out, _, err := httpGet(s.URL, false, "curl/7.26.0")
	if err != nil {
		t.Fatal(err)
	}
	if key := "hostname"; strings.Contains(out, key) {
		t.Errorf("Expected response to not key %q", key)
	}
}

func TestIPFromRequest(t *testing.T) {
	var tests = []struct {
		in  *http.Request
		out net.IP
	}{
		{&http.Request{RemoteAddr: "1.3.3.7:9999"}, net.ParseIP("1.3.3.7")},
		{&http.Request{Header: http.Header{"X-Real-Ip": []string{"1.3.3.7"}}}, net.ParseIP("1.3.3.7")},
	}
	for _, tt := range tests {
		ip, err := ipFromRequest(tt.in)
		if err != nil {
			t.Fatal(err)
		}
		if !ip.Equal(tt.out) {
			t.Errorf("Expected %s, got %s", tt.out, ip)
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
		{"fetch libfetch/2.0", true},
		{"HTTPie/0.9.3", true},
		{browserUserAgent, false},
	}
	for _, tt := range tests {
		r := &http.Request{Header: http.Header{"User-Agent": []string{tt.in}}}
		if got := cliMatcher(r, nil); got != tt.out {
			t.Errorf("Expected %t, got %t for %q", tt.out, got, tt.in)
		}
	}
}
