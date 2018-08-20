package iputil

import (
	"math/big"
	"net"
	"testing"
)

func TestToDecimal(t *testing.T) {
	var msb = new(big.Int)
	msb, _ = msb.SetString("80000000000000000000000000000000", 16)

	var tests = []struct {
		in  string
		out *big.Int
	}{
		{"127.0.0.1", big.NewInt(2130706433)},
		{"::1", big.NewInt(1)},
		{"8000::", msb},
	}
	for _, tt := range tests {
		i := ToDecimal(net.ParseIP(tt.in))
		if tt.out.Cmp(i) != 0 {
			t.Errorf("Expected %d, got %d for IP %s", tt.out, i, tt.in)
		}
	}
}
