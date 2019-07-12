package useragent

import (
	"testing"
)

func TestParse(t *testing.T) {
	var tests = []struct {
		in  string
		out UserAgent
	}{
		{"", UserAgent{}},
		{"curl/", UserAgent{Product: "curl"}},
		{"curl/foo", UserAgent{Product: "curl", Comment: "foo"}},
		{"curl/7.26.0", UserAgent{Product: "curl", Version: "7.26.0"}},
		{"Wget/1.13.4 (linux-gnu)", UserAgent{Product: "Wget", Version: "1.13.4", Comment: "(linux-gnu)"}},
		{"Wget", UserAgent{Product: "Wget"}},
		{"fetch libfetch/2.0", UserAgent{Product: "fetch libfetch", Version: "2.0"}},
		{"Go 1.1 package http", UserAgent{Product: "Go", Comment: "1.1 package http"}},
		{"Mikrotik/6.x Fetch", UserAgent{Product: "Mikrotik", Version: "6.x", Comment: "Fetch"}},
		{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_8_4) " +
			"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/30.0.1599.28 " +
			"Safari/537.36", UserAgent{Product: "Mozilla", Version: "5.0", Comment: "(Macintosh; Intel Mac OS X 10_8_4) " +
			"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/30.0.1599.28 " +
			"Safari/537.36"}},
	}
	for _, tt := range tests {
		ua := Parse(tt.in)
		if got := ua.Product; got != tt.out.Product {
			t.Errorf("got Product=%q for %q, want %q", got, tt.in, tt.out.Product)
		}
		if got := ua.Version; got != tt.out.Version {
			t.Errorf("got Version=%q for %q, want %q", got, tt.in, tt.out.Version)
		}
		if got := ua.Comment; got != tt.out.Comment {
			t.Errorf("got Comment=%q for %q, want %q", got, tt.in, tt.out.Comment)
		}
	}
}
