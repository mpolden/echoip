package database

import (
	"net"

	geoip2 "github.com/oschwald/geoip2-golang"
)

type Client interface {
	Country(net.IP) (Country, error)
	City(net.IP) (string, error)
	IsEmpty() bool
}

type Country struct {
	Name string
	ISO  string
}

type geoip struct {
	country *geoip2.Reader
	city    *geoip2.Reader
}

func New(countryDB, cityDB string) (Client, error) {
	var country, city *geoip2.Reader
	if countryDB != "" {
		r, err := geoip2.Open(countryDB)
		if err != nil {
			return nil, err
		}
		country = r
	}
	if cityDB != "" {
		r, err := geoip2.Open(cityDB)
		if err != nil {
			return nil, err
		}
		city = r
	}
	return &geoip{country: country, city: city}, nil
}

func (g *geoip) Country(ip net.IP) (Country, error) {
	country := Country{}
	if g.country == nil {
		return country, nil
	}
	record, err := g.country.Country(ip)
	if err != nil {
		return country, err
	}
	if c, exists := record.Country.Names["en"]; exists {
		country.Name = c
	}
	if c, exists := record.RegisteredCountry.Names["en"]; exists && country.Name == "" {
		country.Name = c
	}
	if record.Country.IsoCode != "" {
		country.ISO = record.Country.IsoCode
	}
	if record.RegisteredCountry.IsoCode != "" && country.ISO == "" {
		country.ISO = record.RegisteredCountry.IsoCode
	}
	return country, nil
}

func (g *geoip) City(ip net.IP) (string, error) {
	if g.city == nil {
		return "", nil
	}
	record, err := g.city.City(ip)
	if err != nil {
		return "", err
	}
	if city, exists := record.City.Names["en"]; exists {
		return city, nil
	}
	return "", nil
}

func (g *geoip) IsEmpty() bool {
	return g.country == nil && g.city == nil
}
