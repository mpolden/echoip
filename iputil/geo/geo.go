package geo

import (
	"math"
	"net"

	geoip2 "github.com/oschwald/geoip2-golang"
)

type Reader interface {
	Country(net.IP) (Country, error)
	City(net.IP) (City, error)
	ASN(net.IP) (ASN, error)
	IsEmpty() bool
}

type Country struct {
	Name string
	ISO  string
	IsEU *bool
}

type City struct {
	Name      string
	Latitude  float64
	Longitude float64
}

type ASN struct {
	AutonomousSystemNumber       uint
	AutonomousSystemOrganization string
}

type geoip struct {
	country *geoip2.Reader
	city    *geoip2.Reader
	asn     *geoip2.Reader
}

func Open(countryDB, cityDB string, asnDB string) (Reader, error) {
	var country, city, asn *geoip2.Reader
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
	if asnDB != "" {
		r, err := geoip2.Open(asnDB)
		if err != nil {
			return nil, err
		}
		asn = r
	}
	return &geoip{country: country, city: city, asn: asn}, nil
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
	isEU := record.Country.IsInEuropeanUnion || record.RegisteredCountry.IsInEuropeanUnion
	country.IsEU = &isEU
	return country, nil
}

func (g *geoip) City(ip net.IP) (City, error) {
	city := City{}
	if g.city == nil {
		return city, nil
	}
	record, err := g.city.City(ip)
	if err != nil {
		return city, err
	}
	if c, exists := record.City.Names["en"]; exists {
		city.Name = c
	}
	if !math.IsNaN(record.Location.Latitude) {
		city.Latitude = record.Location.Latitude
	}
	if !math.IsNaN(record.Location.Longitude) {
		city.Longitude = record.Location.Longitude
	}
	return city, nil
}

func (g *geoip) ASN(ip net.IP) (ASN, error) {
	asn := ASN{}
	if g.asn == nil {
		return asn, nil
	}
	record, err := g.asn.ASN(ip)
	if err != nil {
		return asn, err
	}
	if record.AutonomousSystemNumber > 0 {
		asn.AutonomousSystemNumber = record.AutonomousSystemNumber
	}
	if record.AutonomousSystemOrganization != "" {
		asn.AutonomousSystemOrganization = record.AutonomousSystemOrganization
	}
	return asn, nil
}

func (g *geoip) IsEmpty() bool {
	return g.country == nil && g.city == nil
}
