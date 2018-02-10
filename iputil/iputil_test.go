package iputil

import (
	"net"
	"testing"
)

func TestToDecimal(t *testing.T) {
	var tests = []struct {
		in  string
		out uint64
	}{
		{"127.0.0.1", 2130706433},
		{"::1", 1},
	}
	for _, tt := range tests {
		i := ToDecimal(net.ParseIP(tt.in))
		if tt.out != i {
			t.Errorf("Expected %d, got %d for IP %s", tt.out, i, tt.in)
		}
	}
}
