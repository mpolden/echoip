package api

import (
	"fmt"
	"net"
	"time"

	"github.com/oschwald/geoip2-golang"
)

type Oracle interface {
	LookupAddr(string) ([]string, error)
	LookupCountry(net.IP) (string, error)
	LookupCity(net.IP) (string, error)
	LookupPort(net.IP, uint64) error
	IsLookupAddrEnabled() bool
	IsLookupCountryEnabled() bool
	IsLookupCityEnabled() bool
	IsLookupPortEnabled() bool
}

type DefaultOracle struct {
	lookupAddr           func(string) ([]string, error)
	lookupCountry        func(net.IP) (string, error)
	lookupCity           func(net.IP) (string, error)
	lookupPort           func(net.IP, uint64) error
	lookupAddrEnabled    bool
	lookupCountryEnabled bool
	lookupCityEnabled    bool
	lookupPortEnabled    bool
}

func NewOracle() *DefaultOracle {
	return &DefaultOracle{
		lookupAddr:    func(string) ([]string, error) { return nil, nil },
		lookupCountry: func(net.IP) (string, error) { return "", nil },
		lookupCity:    func(net.IP) (string, error) { return "", nil },
		lookupPort:    func(net.IP, uint64) error { return nil },
	}
}

func (r *DefaultOracle) LookupAddr(address string) ([]string, error) {
	return r.lookupAddr(address)
}

func (r *DefaultOracle) LookupCountry(ip net.IP) (string, error) {
	return r.lookupCountry(ip)
}

func (r *DefaultOracle) LookupCity(ip net.IP) (string, error) {
	return r.lookupCity(ip)
}

func (r *DefaultOracle) LookupPort(ip net.IP, port uint64) error {
	return r.lookupPort(ip, port)
}

func (r *DefaultOracle) EnableLookupAddr() {
	r.lookupAddr = net.LookupAddr
	r.lookupAddrEnabled = true
}

func (r *DefaultOracle) EnableLookupCountry(filepath string) error {
	db, err := geoip2.Open(filepath)
	if err != nil {
		return err
	}
	r.lookupCountry = func(ip net.IP) (string, error) {
		return lookupCountry(db, ip)
	}
	r.lookupCountryEnabled = true
	return nil
}

func (r *DefaultOracle) EnableLookupCity(filepath string) error {
	db, err := geoip2.Open(filepath)
	if err != nil {
		return err
	}
	r.lookupCity = func(ip net.IP) (string, error) {
		return lookupCity(db, ip)
	}
	r.lookupCityEnabled = true
	return nil
}

func (r *DefaultOracle) EnableLookupPort() {
	r.lookupPort = lookupPort
	r.lookupPortEnabled = true
}

func (r *DefaultOracle) IsLookupAddrEnabled() bool    { return r.lookupAddrEnabled }
func (r *DefaultOracle) IsLookupCountryEnabled() bool { return r.lookupCountryEnabled }
func (r *DefaultOracle) IsLookupCityEnabled() bool    { return r.lookupCityEnabled }
func (r *DefaultOracle) IsLookupPortEnabled() bool    { return r.lookupPortEnabled }

func lookupPort(ip net.IP, port uint64) error {
	address := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("tcp", address, 2*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()
	return nil
}

func lookupCountry(db *geoip2.Reader, ip net.IP) (string, error) {
	record, err := db.Country(ip)
	if err != nil {
		return "", err
	}
	if country, exists := record.Country.Names["en"]; exists {
		return country, nil
	}
	if country, exists := record.RegisteredCountry.Names["en"]; exists {
		return country, nil
	}
	return "Unknown", fmt.Errorf("could not determine country for IP: %s", ip)
}

func lookupCity(db *geoip2.Reader, ip net.IP) (string, error) {
	record, err := db.City(ip)
	if err != nil {
		return "", err
	}
	if city, exists := record.City.Names["en"]; exists {
		return city, nil
	}
	return "Unknown", fmt.Errorf("could not determine city for IP: %s", ip)
}
