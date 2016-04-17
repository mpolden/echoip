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
		ReverseLookup bool   `short:"r" long:"reverse-lookup" description:"Perform reverse hostname lookups"`
		PortLookup    bool   `short:"p" long:"port-lookup" description:"Enable port lookup"`
		Template      string `short:"t" long:"template" description:"Path to template" default:"index.html"`
		IPHeader      string `short:"H" long:"trusted-header" description:"Header to trust for remote IP, if present (e.g. X-Real-IP)"`
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
	if opts.IPHeader != "" {
		log.Printf("Trusting header %s to contain correct remote IP", opts.IPHeader)
	}

	api := api.New(oracle)
	api.Template = opts.Template
	api.IPHeader = opts.IPHeader

	log.Printf("Listening on %s", opts.Listen)
	if err := api.ListenAndServe(opts.Listen); err != nil {
		log.Fatal(err)
	}
}
