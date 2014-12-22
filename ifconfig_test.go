package main

import (
	"net/url"
	"testing"
)

func TestIsCLI(t *testing.T) {
	userAgents := []string{"curl/7.26.0", "Wget/1.13.4 (linux-gnu)",
		"fetch libfetch/2.0"}

	for _, userAgent := range userAgents {
		if !isCLI(userAgent) {
			t.Errorf("Expected true for %s", userAgent)
		}
	}

	browserUserAgent := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_8_4) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/30.0.1599.28 " +
		"Safari/537.36"
	if isCLI(browserUserAgent) {
		t.Errorf("Expected false for %s", browserUserAgent)
	}
}

func TestPathToKey(t *testing.T) {
	if key := pathToKey("/ip"); key != "ip" {
		t.Fatalf("Expected 'ip', got '%s'", key)
	}
	if key := pathToKey("/User-Agent"); key != "user-agent" {
		t.Fatalf("Expected 'user-agent', got '%s'", key)
	}
	if key := pathToKey("/all.json"); key != "all" {
		t.Fatalf("Expected 'all', got '%s'", key)
	}
}

func TestLookupCmd(t *testing.T) {
	values := url.Values{"cmd": []string{"curl"}}
	if v := lookupCmd(values); v.Name != "curl" {
		t.Fatalf("Expected 'curl', got '%s'", v)
	}
	values = url.Values{"cmd": []string{"foo"}}
	if v := lookupCmd(values); v.Name != "curl" {
		t.Fatalf("Expected 'curl', got '%s'", v)
	}
	values = url.Values{}
	if v := lookupCmd(values); v.Name != "curl" {
		t.Fatalf("Expected 'curl', got '%s'", v)
	}
	values = url.Values{"cmd": []string{"wget"}}
	if v := lookupCmd(values); v.Name != "wget" {
		t.Fatalf("Expected 'wget', got '%s'", v)
	}
	values = url.Values{"cmd": []string{"fetch"}}
	if v := lookupCmd(values); v.Name != "fetch" {
		t.Fatalf("Expected 'fetch', got '%s'", v)
	}
}
