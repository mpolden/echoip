package useragent

import (
	"strings"
)

type UserAgent struct {
	Product string
	Version string
	Comment string
}

func Parse(s string) UserAgent {
	parts := strings.SplitN(s, "/", 2)
	var version, comment string
	if len(parts) > 1 {
		// If first character is a number, treat it as version
		if parts[1][0] >= 48 && parts[1][0] <= 57 {
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
	return UserAgent{
		Product: parts[0],
		Version: version,
		Comment: comment,
	}
}
