package main

import (
	"io"
	"log"
	"strings"

	"os"

	"github.com/BurntSushi/toml"
	"github.com/levelsoftware/echoip/cache"
	"github.com/levelsoftware/echoip/http"
	"github.com/levelsoftware/echoip/iputil"
	"github.com/levelsoftware/echoip/iputil/geo"
	"github.com/levelsoftware/echoip/iputil/ipstack"
	parser "github.com/levelsoftware/echoip/iputil/paser"
	ipstackApi "github.com/qioalice/ipstack"
)

type multiValueFlag []string

func (f *multiValueFlag) String() string {
	return strings.Join([]string(*f), ", ")
}

func (f *multiValueFlag) Set(v string) error {
	*f = append(*f, v)
	return nil
}

func init() {
	log.SetPrefix("echoip: ")
	log.SetFlags(log.Lshortfile)
}

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

func main() {
	file, err := os.Open("/etc/echoip/config.toml")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var config Config

	b, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	err = toml.Unmarshal(b, &config)
	if err != nil {
		panic(err)
	}

	var parser parser.Parser
	if config.Database == "geoip" {
		log.Print("Using GeoIP for IP database")
		geo, err := geo.Open(
			config.GeoIP.CountryFile,
			config.GeoIP.CityFile,
			config.GeoIP.AsnFile,
		)
		if err != nil {
			log.Fatal(err)
		}
		parser = &geo
	}

	if config.Database == "ipstack" {
		log.Print("Using GeoIP for IP database")
		if config.IPStack.EnableSecurity {
			log.Print("Enable Security Module ( Requires Professional Plus account )")
		}
		enableSecurity := ipstackApi.ParamEnableSecurity(config.IPStack.EnableSecurity)
		apiKey := ipstackApi.ParamToken(config.IPStack.ApiKey)
		useHttps := ipstackApi.ParamUseHTTPS(config.IPStack.UseHttps)
		if config.IPStack.UseHttps {
			log.Print("Use IP Stack HTTPS API ( Requires non-free account )")
		}
		if err := ipstackApi.Init(apiKey, enableSecurity, useHttps); err != nil {
			log.Fatal(err)
		}
		ips := ipstack.IPStack{}
		parser = &ips
	}

	var serverCache cache.Cache
	if len(config.RedisUrl) > 0 {
		redisCache, err := cache.RedisCache(config.RedisUrl)
		serverCache = &redisCache
		if err != nil {
			log.Fatal(err)
		}
	} else {
		serverCache = &cache.Null{}
	}

	server := http.New(parser, serverCache, config.CacheTtl, config.Profile)
	server.IPHeaders = config.TrustedHeaders

	if _, err := os.Stat(config.TemplateDir); err == nil {
		server.Template = config.TemplateDir
	} else {
		log.Printf("Not configuring default handler: Template not found: %s", config.TemplateDir)
	}
	if config.ReverseLookup {
		log.Println("Enabling reverse lookup")
		server.LookupAddr = iputil.LookupAddr
	}
	if config.PortLookup {
		log.Println("Enabling port lookup")
		server.LookupPort = iputil.LookupPort
	}
	if config.ShowSponsor {
		log.Println("Enabling sponsor logo")
		server.Sponsor = config.ShowSponsor
	}
	if len(config.TrustedHeaders) > 0 {
		log.Printf("Trusting remote IP from header(s): %s", config.TrustedHeaders)
	}
	if config.Profile {
		log.Printf("Enabling profiling handlers")
	}
	log.Printf("Listening on http://%s", config.Listen)
	if err := server.ListenAndServe(config.Listen); err != nil {
		log.Fatal(err)
	}
}
