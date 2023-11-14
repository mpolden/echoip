package ipstack

import (
	"fmt"
	"net"
	"time"

	"github.com/levelsoftware/echoip/iputil"
	parser "github.com/levelsoftware/echoip/iputil/paser"

	"github.com/qioalice/ipstack"
)

type IPStack struct {
	response *ipstack.Response
}

func (ips *IPStack) Parse(ip net.IP, hostname string) (parser.Response, error) {
	res, err := ipstack.IP(ip.String())
	ips.response = res

	if err != nil {
		return parser.Response{}, err
	}

	ipDecimal := iputil.ToDecimal(ip)

	parserResponse := parser.Response{
		UsingGeoIP:   false,
		UsingIPStack: true,

		/* kept for backward compatibility */
		Latitude:   float64(res.Latitide),
		Longitude:  float64(res.Longitude),
		Hostname:   hostname,
		IP:         ip,
		IPDecimal:  ipDecimal,
		Country:    res.CountryName,
		CountryISO: res.CountryCode,
		RegionName: res.RegionName,
		RegionCode: res.RegionCode,
		MetroCode:  0,
		PostalCode: res.Zip,
		City:       res.City,
	}

	ips.ParseSecurityResponse(&parserResponse)
	ips.ParseTimezoneResponse(&parserResponse)
	ips.ParseLocationResponse(&parserResponse)
	ips.ParseConnectionResponse(&parserResponse)
	ips.ParseCurrencyResponse(&parserResponse)

	return parserResponse, nil
}

func (ips *IPStack) ParseSecurityResponse(parserResponse *parser.Response) {
	if ips.response.Security != nil {
		parserResponse.IPStackSecurityEnabled = true

		parserResponse.Security = parser.Security{
			IsProxy:     ips.response.Security.IsProxy,
			IsTor:       ips.response.Security.IsTOR,
			CrawlerName: ips.response.Security.CrawlerName,
			CrawlerType: ips.response.Security.CrawlerType,
			ThreatLevel: ips.response.Security.ThreatLevel,
		}

		if threat_types, ok := ips.response.Security.ThreatTypes.([]string); ok {
			parserResponse.Security.ThreatTypes = threat_types
		}
	}
}

func (ips *IPStack) ParseTimezoneResponse(parserResponse *parser.Response) {
	if ips.response.Timezone != nil {
		parserResponse.TimezoneEtc = parser.Timezone{
			ID:                ips.response.Timezone.ID,
			CurrentTime:       ips.response.Timezone.CurrentTime.Format(time.RFC3339),
			GmtOffset:         ips.response.Timezone.GMTOffset,
			Code:              ips.response.Timezone.Code,
			IsDaylightSavings: ips.response.Timezone.IsDaylightSaving,
		}

		/* kept for backward compatibility */
		parserResponse.Timezone = ips.response.Timezone.ID
	}
}

func (ips *IPStack) ParseLocationResponse(parserResponse *parser.Response) {
	if ips.response.Location != nil {
		var languages []parser.Language
		for i := 0; i < len(ips.response.Location.Languages); i++ {
			languages = append(languages, parser.Language{
				Code:   ips.response.Location.Languages[i].Code,
				Name:   ips.response.Location.Languages[i].Name,
				Native: ips.response.Location.Languages[i].NativeName,
			})
		}
		parserResponse.Location = parser.Location{
			Languages: languages,
			CountryFlag: parser.CountryFlag{
				Flag:         ips.response.Location.CountryFlagLink,
				Emoji:        ips.response.Location.CountryFlagEmoji,
				EmojiUnicode: ips.response.Location.CountryFlagEmojiUnicode,
			},
		}

		/* kept for backward compatibility */
		parserResponse.CountryEU = &ips.response.Location.IsEU
	}
}

func (ips *IPStack) ParseConnectionResponse(parserResponse *parser.Response) {
	if ips.response.Connection != nil && ips.response.Connection.ASN > 0 {
		/* kept for backward compatibility */
		parserResponse.ASN = fmt.Sprintf("AS%d", ips.response.Connection.ASN)
	}
}

func (ips *IPStack) ParseCurrencyResponse(parserResponse *parser.Response) {
	if ips.response.Currency != nil {
		parserResponse.Currency = parser.Currency{
			Code:         parserResponse.Currency.Code,
			Name:         parserResponse.Currency.Name,
			Plural:       parserResponse.Currency.Plural,
			Symbol:       parserResponse.Currency.Symbol,
			SymbolNative: parserResponse.Currency.SymbolNative,
		}
	}
}

func (ips *IPStack) IsEmpty() bool {
	return false
}
