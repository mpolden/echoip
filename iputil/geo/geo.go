package geo

import (
	"fmt"
	"math"
	"net"

	"github.com/mpolden/echoip/iputil"
	parser "github.com/mpolden/echoip/iputil/paser"
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
	Name       string
	Latitude   float64
	Longitude  float64
	PostalCode string
	Timezone   string
	MetroCode  uint
	RegionName string
	RegionCode string
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

func Open(countryDB, cityDB string, asnDB string) (geoip, error) {
	var country, city, asn *geoip2.Reader
	if countryDB != "" {
		r, err := geoip2.Open(countryDB)
		if err != nil {
			return geoip{}, err
		}
		country = r
	}
	if cityDB != "" {
		r, err := geoip2.Open(cityDB)
		if err != nil {
			return geoip{}, err
		}
		city = r
	}
	if asnDB != "" {
		r, err := geoip2.Open(asnDB)
		if err != nil {
			return geoip{}, err
		}
		asn = r
	}
	return geoip{country: country, city: city, asn: asn}, nil
}

func (g *geoip) Parse(ip net.IP, hostname string) (parser.Response, error) {
	ipDecimal := iputil.ToDecimal(ip)
	country, _ := g.Country(ip)
	city, _ := g.City(ip)
	asn, _ := g.ASN(ip)
	var autonomousSystemNumber string
	if asn.AutonomousSystemNumber > 0 {
		autonomousSystemNumber = fmt.Sprintf("AS%d", asn.AutonomousSystemNumber)
	}
	return parser.Response{
		UsingGeoIP:   true,
		UsingIPStack: false,
		IP:           ip,
		IPDecimal:    ipDecimal,
		Country:      country.Name,
		CountryISO:   country.ISO,
		CountryEU:    country.IsEU,
		RegionName:   city.RegionName,
		RegionCode:   city.RegionCode,
		MetroCode:    city.MetroCode,
		PostalCode:   city.PostalCode,
		City:         city.Name,
		Latitude:     city.Latitude,
		Longitude:    city.Longitude,
		Timezone:     city.Timezone,
		ASN:          autonomousSystemNumber,
		ASNOrg:       asn.AutonomousSystemOrganization,
		Hostname:     hostname,
	}, nil
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
	if len(record.Subdivisions) > 0 {
		if c, exists := record.Subdivisions[0].Names["en"]; exists {
			city.RegionName = c
		}
		if record.Subdivisions[0].IsoCode != "" {
			city.RegionCode = record.Subdivisions[0].IsoCode
		}
	}
	if !math.IsNaN(record.Location.Latitude) {
		city.Latitude = record.Location.Latitude
	}
	if !math.IsNaN(record.Location.Longitude) {
		city.Longitude = record.Location.Longitude
	}
	// Metro code is US Only https://maxmind.github.io/GeoIP2-dotnet/doc/v2.7.1/html/P_MaxMind_GeoIP2_Model_Location_MetroCode.htm
	if record.Location.MetroCode > 0 && record.Country.IsoCode == "US" {
		city.MetroCode = record.Location.MetroCode
	}
	if record.Postal.Code != "" {
		city.PostalCode = record.Postal.Code
	}
	if record.Location.TimeZone != "" {
		city.Timezone = record.Location.TimeZone
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
