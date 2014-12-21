package main

import "testing"

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
