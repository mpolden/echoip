package ipstack

import (
	"fmt"
	"net"

	"github.com/mpolden/echoip/iputil"
	parser "github.com/mpolden/echoip/iputil/paser"
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
	fmt.Printf("%+v\n", res)

	ipDecimal := iputil.ToDecimal(ip)

	parserResponse := parser.Response{
		UsingGeoIP:   false,
		UsingIPStack: true,
		Latitude:     float64(res.Latitide),
		Longitude:    float64(res.Longitude),
		Hostname:     hostname,
		IP:           ip,
		IPDecimal:    ipDecimal,
		Country:      res.CountryName,
		CountryISO:   res.CountryCode,
		RegionName:   res.RegionName,
		RegionCode:   res.RegionCode,
		MetroCode:    0,
		PostalCode:   res.Zip,
		City:         res.City,
	}

	if res.Timezone != nil {
		parserResponse.Timezone = res.Timezone.ID
	}

	if res.Location != nil {
		parserResponse.CountryEU = &res.Location.IsEU
	}

	if res.Connection != nil {
		if res.Connection.ASN > 0 {
			parserResponse.ASN = fmt.Sprintf("AS%d", res.Connection.ASN)
		}
	}

	return parserResponse, nil
}

func (ips *IPStack) IsEmpty() bool {
	return false
}
