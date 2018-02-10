package iputil

import (
	"fmt"
	"math/big"
	"net"
	"strings"
	"time"
)

func LookupAddr(ip net.IP) ([]string, error) {
	names, err := net.LookupAddr(ip.String())
	for i, _ := range names {
		names[i] = strings.TrimRight(names[i], ".") // Always return unrooted name
	}
	return names, err
}

func LookupPort(ip net.IP, port uint64) error {
	address := fmt.Sprintf("[%s]:%d", ip, port)
	conn, err := net.DialTimeout("tcp", address, 2*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()
	return nil
}

func ToDecimal(ip net.IP) *big.Int {
	i := big.NewInt(0)
	if to4 := ip.To4(); to4 != nil {
		i.SetBytes(to4)
	} else {
		i.SetBytes(ip)
	}
	return i
}
