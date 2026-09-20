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
		// PowerShell identifies as Mozilla and appends its own product last, on
		// every platform. Windows PowerShell 5.1 and PowerShell 6+ are separate
		// products, so they are reported separately. The comment may contain
		// nested parens, so the product is found from the end of the string.
		{"Mozilla/5.0 (Windows NT; Windows NT 10.0; en-US) WindowsPowerShell/5.1.19041.1",
			UserAgent{Product: "WindowsPowerShell", Version: "5.1.19041.1",
				Comment: "(Windows NT; Windows NT 10.0; en-US)"}},
		{"Mozilla/5.0 (Windows NT 10.0; Microsoft Windows 10.0.19045; en-US) PowerShell/7.4.6",
			UserAgent{Product: "PowerShell", Version: "7.4.6",
				Comment: "(Windows NT 10.0; Microsoft Windows 10.0.19045; en-US)"}},
		{"Mozilla/5.0 (Linux; Fedora Linux 44 (WSL); en-US) PowerShell/7.6.5",
			UserAgent{Product: "PowerShell", Version: "7.6.5",
				Comment: "(Linux; Fedora Linux 44 (WSL); en-US)"}},
		{"Mozilla/5.0 (Macintosh; Darwin 23.6.0; en-US) PowerShell/7.5.0",
			UserAgent{Product: "PowerShell", Version: "7.5.0",
				Comment: "(Macintosh; Darwin 23.6.0; en-US)"}},
		// The leading space keeps an unrelated product ending in PowerShell from
		// matching, and keeps PowerShell from swallowing WindowsPowerShell.
		{"Mozilla/5.0 (Windows NT 10.0; en-US) NotPowerShell/1.0",
			UserAgent{Product: "Mozilla", Version: "5.0",
				Comment: "(Windows NT 10.0; en-US) NotPowerShell/1.0"}},
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
