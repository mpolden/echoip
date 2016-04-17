package main

import (
	flags "github.com/jessevdk/go-flags"

	"log"
	"os"

	"github.com/martinp/ipd/api"
)

func main() {
	var opts struct {
		CountryDBPath string `short:"f" long:"country-db" description:"Path to GeoIP country database" value-name:"FILE" default:""`
		CityDBPath    string `short:"c" long:"city-db" description:"Path to GeoIP city database" value-name:"FILE" default:""`
		Listen        string `short:"l" long:"listen" description:"Listening address" value-name:"ADDR" default:":8080"`
		CORS          bool   `short:"x" long:"cors" description:"Allow requests from other domains"`
		ReverseLookup bool   `short:"r" long:"reverse-lookup" description:"Perform reverse hostname lookups"`
		PortLookup    bool   `short:"p" long:"port-lookup" description:"Enable port lookup"`
		Template      string `short:"t" long:"template" description:"Path to template" default:"index.html"`
	}
	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		os.Exit(1)
	}

	oracle := api.NewOracle()
	if opts.ReverseLookup {
		log.Println("Enabling reverse lookup")
		oracle.EnableLookupAddr()
	}
	if opts.PortLookup {
		log.Println("Enabling port lookup")
		oracle.EnableLookupPort()
	}
	if opts.CountryDBPath != "" {
		log.Printf("Enabling country lookup (using database: %s)", opts.CountryDBPath)
		if err := oracle.EnableLookupCountry(opts.CountryDBPath); err != nil {
			log.Fatal(err)
		}
	}
	if opts.CityDBPath != "" {
		log.Printf("Enabling city lookup (using database: %s)", opts.CityDBPath)
		if err := oracle.EnableLookupCity(opts.CityDBPath); err != nil {
			log.Fatal(err)
		}
	}

	api := api.New(oracle)
	api.CORS = opts.CORS
	api.Template = opts.Template

	log.Printf("Listening on %s", opts.Listen)
	if err := api.ListenAndServe(opts.Listen); err != nil {
		log.Fatal(err)
	}
}
