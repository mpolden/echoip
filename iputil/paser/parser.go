package parser

import (
	"math/big"
	"net"

	"github.com/mpolden/echoip/useragent"
)

type Parser interface {
	Parse(net.IP, string) (Response, error)
	IsEmpty() bool
}

type Response struct {
	UsingGeoIP   bool                 `json:"UsingGeoIP"`
	UsingIPStack bool                 `json:"UsingIPStack"`
	IP           net.IP               `json:"ip"`
	IPDecimal    *big.Int             `json:"ip_decimal"`
	Country      string               `json:"country,omitempty"`
	CountryISO   string               `json:"country_iso,omitempty"`
	CountryEU    *bool                `json:"country_eu,omitempty"`
	RegionName   string               `json:"region_name,omitempty"`
	RegionCode   string               `json:"region_code,omitempty"`
	MetroCode    uint                 `json:"metro_code,omitempty"`
	PostalCode   string               `json:"zip_code,omitempty"`
	City         string               `json:"city,omitempty"`
	Latitude     float64              `json:"latitude,omitempty"`
	Longitude    float64              `json:"longitude,omitempty"`
	Timezone     string               `json:"time_zone,omitempty"`
	ASN          string               `json:"asn,omitempty"`
	ASNOrg       string               `json:"asn_org,omitempty"`
	Hostname     string               `json:"hostname,omitempty"`
	UserAgent    *useragent.UserAgent `json:"user_agent,omitempty"`
}
