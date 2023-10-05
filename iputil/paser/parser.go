package parser

import (
	"math/big"
	"net"

	"github.com/levelsoftware/echoip/useragent"
)

type Parser interface {
	Parse(net.IP, string) (Response, error)
	IsEmpty() bool
}

type Currency struct {
	Code         string `json:"code,omitempty"`
	Name         string `json:"name,omitempty"`
	Plural       string `json:"plural,omitempty"`
	Symbol       string `json:"symbol,omitempty"`
	SymbolNative string `json:"symbol_native,omitempty"`
}

type Security struct {
	IsProxy     bool     `json:"is_proxy"`
	IsCrawler   bool     `json:"is_crawler"`
	CrawlerName string   `json:"crawler_name,omitempty"`
	CrawlerType string   `json:"crawler_type,omitempty"`
	IsTor       bool     `json:"is_tor"`
	ThreatLevel string   `json:"threat_level,omitempty"`
	ThreatTypes []string `json:"threat_types,omitempty"`
}

type Timezone struct {
	ID                string `json:"id,omitempty"`
	CurrentTime       string `json:"current_time,omitempty"`
	GmtOffset         int    `json:"gmt_offset,omitempty"`
	Code              string `json:"code,omitempty"`
	IsDaylightSavings bool   `json:"is_daylight_savings,omitempty"`
}

type Language struct {
	Code   string `json:"code,omitempty"`
	Name   string `json:"name,omitempty"`
	Native string `json:"native,omitempty"`
}

type CountryFlag struct {
	Flag         string `json:"flag,omitempty"`
	Emoji        string `json:"emoji,omitempty"`
	EmojiUnicode string `json:"emoji_unicode,omitempty"`
}

type Location struct {
	Languages   interface{} `json:"languages,omitempty"`
	CountryFlag CountryFlag `json:"country_flag,omitempty"`
}

type Response struct {
	UsingGeoIP             bool `json:"UsingGeoIP"`
	UsingIPStack           bool `json:"UsingIPStack"`
	IPStackSecurityEnabled bool `json:"IPStackSecurityEnabled"`

	TimezoneEtc Timezone `json:"timezone_etc,omitempty"`
	Security    Security `json:"security,omitempty"`
	Currency    Currency `json:"currency,omitempty"`
	Location    Location `json:"location,omitempty"`

	/* Kept to prevent breaking changes */
	IP         net.IP               `json:"ip"`
	IPDecimal  *big.Int             `json:"ip_decimal"`
	Country    string               `json:"country,omitempty"`
	CountryISO string               `json:"country_iso,omitempty"`
	CountryEU  *bool                `json:"country_eu,omitempty"`
	RegionName string               `json:"region_name,omitempty"`
	RegionCode string               `json:"region_code,omitempty"`
	MetroCode  uint                 `json:"metro_code,omitempty"`
	PostalCode string               `json:"zip_code,omitempty"`
	City       string               `json:"city,omitempty"`
	Latitude   float64              `json:"latitude,omitempty"`
	Longitude  float64              `json:"longitude,omitempty"`
	Timezone   string               `json:"timezone,omitempty"`
	ASN        string               `json:"asn,omitempty"`
	ASNOrg     string               `json:"asn_org,omitempty"`
	Hostname   string               `json:"hostname,omitempty"`
	UserAgent  *useragent.UserAgent `json:"user_agent,omitempty"`
}
