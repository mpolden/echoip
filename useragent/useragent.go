package useragent

import (
	"strings"
)

type UserAgent struct {
	Product  string `json:"product,omitempty"`
	Version  string `json:"version,omitempty"`
	Comment  string `json:"comment,omitempty"`
	RawValue string `json:"raw_value,omitempty"`
}

func Parse(s string) UserAgent {
	parts := strings.SplitN(s, "/", 2)
	var version, comment string
	if len(parts) > 1 {
		// If first character is a number, treat it as version
		if len(parts[1]) > 0 && parts[1][0] >= 48 && parts[1][0] <= 57 {
			rest := strings.SplitN(parts[1], " ", 2)
			version = rest[0]
			if len(rest) > 1 {
				comment = rest[1]
			}
		} else {
			comment = parts[1]
		}
	} else {
		parts = strings.SplitN(s, " ", 2)
		if len(parts) > 1 {
			comment = parts[1]
		}
	}
	return parsePowerShell(UserAgent{
		Product:  parts[0],
		Version:  version,
		Comment:  comment,
		RawValue: s,
	})
}

// PowerShell's Invoke-WebRequest and Invoke-RestMethod identify as Mozilla and
// append their own product last. Windows PowerShell and PowerShell are separate
// products.
func parsePowerShell(ua UserAgent) UserAgent {
	if ua.Product != "Mozilla" {
		return ua
	}
	for _, product := range []string{"WindowsPowerShell", "PowerShell"} {
		comment, version, found := strings.Cut(ua.Comment, " "+product+"/")
		if !found {
			continue
		}
		ua.Product = product
		ua.Version, _, _ = strings.Cut(version, " ")
		ua.Comment = comment
		return ua
	}
	return ua
}
