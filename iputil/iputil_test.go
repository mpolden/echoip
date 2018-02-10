package iputil

import (
	"math/big"
	"net"
	"testing"
)

func TestToDecimal(t *testing.T) {
	var tests = []struct {
		in  string
		out *big.Int
	}{
		{"127.0.0.1", big.NewInt(2130706433)},
		{"::1", big.NewInt(1)},
	}
	for _, tt := range tests {
		i := ToDecimal(net.ParseIP(tt.in))
		if i.Cmp(tt.out) != 0 {
			t.Errorf("Expected %d, got %d for IP %s", tt.out, i, tt.in)
		}
	}
}
