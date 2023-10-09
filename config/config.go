package config

import (
	"os"
	"strconv"
	"strings"
)

type IPStack struct {
	ApiKey         string
	UseHttps       bool
	EnableSecurity bool
}

type GeoIP struct {
	CountryFile string
	CityFile    string
	AsnFile     string
}

type Config struct {
	Listen         string
	TemplateDir    string
	RedisUrl       string
	CacheTtl       int
	ReverseLookup  bool
	PortLookup     bool
	ShowSponsor    bool
	TrustedHeaders []string

	Database string
	Profile  bool

	IPStack IPStack
	GeoIP   GeoIP
}

func GetConfig() (Config, error) {
	defaultConfig := Config{
		Listen:      getenv_string("ECHOIP_LISTEN", ":8080"),
		TemplateDir: getenv_string("ECHOIP_TEMPLATE_DIR", "html/"),
		RedisUrl:    getenv_string("ECHOIP_REDIS_URL", ""),
		Database:    getenv_string("ECHOIP_DATABASE", "geoip"),
		IPStack: IPStack{
			ApiKey: getenv_string("ECHOIP_IPSTACK_API_KEY", ""),
		},
		GeoIP: GeoIP{
			CountryFile: getenv_string("ECHOIP_GEOIP_COUNTRY_FILE", ""),
			CityFile:    getenv_string("ECHOIP_GEOIP_CITY_FILE", ""),
			AsnFile:     getenv_string("ECHOIP_GEOIP_ASN_FILE", ""),
		},
	}

	cacheTtl, err := getenv_int("ECHOIP_CACHE_TTL", 3600)
	if err != nil {
		return Config{}, err
	}
	defaultConfig.CacheTtl = cacheTtl

	reverseLookup, err := getenv_bool("ECHOIP_REVERSE_LOOKUP", false)
	if err != nil {
		return Config{}, err
	}
	defaultConfig.ReverseLookup = reverseLookup

	portLookup, err := getenv_bool("ECHOIP_PORT_LOOKUP", false)
	if err != nil {
		return Config{}, err
	}
	defaultConfig.PortLookup = portLookup

	showSponsor, err := getenv_bool("ECHOIP_SHOW_SPONSOR", false)
	if err != nil {
		return Config{}, err
	}
	defaultConfig.ShowSponsor = showSponsor

	profile, err := getenv_bool("ECHOIP_PROFILE", false)
	if err != nil {
		return Config{}, err
	}
	defaultConfig.Profile = profile

	ipStackUseHttps, err := getenv_bool("ECHOIP_IPSTACK_USE_HTTPS", false)
	if err != nil {
		return Config{}, err
	}
	defaultConfig.IPStack.UseHttps = ipStackUseHttps

	ipStackEnableSecurity, err := getenv_bool("ECHOIP_IPSTACK_ENABLE_SECURITY", false)
	if err != nil {
		return Config{}, err
	}
	defaultConfig.IPStack.EnableSecurity = ipStackEnableSecurity

	trustedHeaders := getenv_string("ECHOIP_TRUSTED_HEADERS", "")
	defaultConfig.TrustedHeaders = strings.Split(trustedHeaders, ",")

	return defaultConfig, nil
}

func getenv_int(key string, fallback int) (int, error) {
	value := os.Getenv(key)

	if len(value) > 0 {
		intValue, err := strconv.Atoi(value)
		if err != nil {
			return 0, err
		}

		return intValue, nil
	}

	return fallback, nil
}

func getenv_string(key string, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func getenv_bool(key string, fallback bool) (bool, error) {
	value := os.Getenv(key)

	if len(value) > 0 {
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return false, err
		}

		return boolValue, nil
	}

	return fallback, nil
}
