package main

import (
	"io"
	"log"
	"strings"

	"os"

	"github.com/BurntSushi/toml"
	"github.com/levelsoftware/echoip/cache"
	"github.com/levelsoftware/echoip/config"
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

func main() {
	runConfig, err := config.GetConfig()

	if err != nil {
		log.Fatalf("Error building configuration: %s", err)
	}

	file, err := os.Open("/etc/echoip/config.toml")
	defer file.Close()

	if err != nil {
		log.Printf("Error opening config file (/etc/echoip/config.toml): %s", err)
	}

	if err == nil {
		var b []byte
		b, err = io.ReadAll(file)
		if err != nil {
			log.Printf("Error reading config file (/etc/echoip/config.toml): %s", err)
		}

		err = toml.Unmarshal(b, &runConfig)
		if err != nil {
			log.Fatalf("Error parsing config file: %s", err)
		}
	}

	var parser parser.Parser
	if runConfig.Database == "geoip" {
		log.Print("Using GeoIP for IP database")
		geo, err := geo.Open(
			runConfig.GeoIP.CountryFile,
			runConfig.GeoIP.CityFile,
			runConfig.GeoIP.AsnFile,
		)
		if err != nil {
			log.Fatal(err)
		}
		parser = &geo
	}

	if runConfig.Database == "ipstack" {
		log.Print("Using IP Stack for IP database")
		if runConfig.IPStack.EnableSecurity {
			log.Print("Enable IP Stack Security Module ( Requires Professional Plus account )")
		}
		enableSecurity := ipstackApi.ParamEnableSecurity(runConfig.IPStack.EnableSecurity)
		apiKey := ipstackApi.ParamToken(runConfig.IPStack.ApiKey)
		useHttps := ipstackApi.ParamUseHTTPS(runConfig.IPStack.UseHttps)
		if runConfig.IPStack.UseHttps {
			log.Print("Use IP Stack HTTPS API ( Requires non-free account )")
		}
		if err := ipstackApi.Init(apiKey, enableSecurity, useHttps); err != nil {
			log.Fatalf("Error initializing IP Stack client: %s", err)
		}
		ips := ipstack.IPStack{}
		parser = &ips
	}

	var serverCache cache.Cache
	if len(runConfig.RedisUrl) > 0 {
		redisCache, err := cache.RedisCache(runConfig.RedisUrl)
		serverCache = &redisCache
		if err != nil {
			log.Fatal(err)
		}
	} else {
		serverCache = &cache.Null{}
	}

	server := http.New(parser, serverCache, runConfig.CacheTtl, runConfig.Profile)
	server.IPHeaders = runConfig.TrustedHeaders

	if _, err := os.Stat(runConfig.TemplateDir); err == nil {
		server.Template = runConfig.TemplateDir
	} else {
		log.Printf("Not configuring default handler: Template not found: %s", runConfig.TemplateDir)
	}
	if runConfig.ReverseLookup {
		log.Println("Enabling reverse lookup")
		server.LookupAddr = iputil.LookupAddr
	}
	if runConfig.PortLookup {
		log.Println("Enabling port lookup")
		server.LookupPort = iputil.LookupPort
	}
	if runConfig.ShowSponsor {
		log.Println("Enabling sponsor logo")
		server.Sponsor = runConfig.ShowSponsor
	}
	if len(runConfig.TrustedHeaders) > 0 {
		log.Printf("Trusting remote IP from header(s): %s", runConfig.TrustedHeaders)
	}
	if runConfig.Profile {
		log.Printf("Enabling profiling handlers")
	}

	log.Printf("Listening on http://%s", runConfig.Listen)
	if err := server.ListenAndServe(runConfig.Listen); err != nil {
		log.Fatal(err)
	}
}
