package main

import (
	flags "github.com/jessevdk/go-flags"

	"log"
	"os"

	"github.com/martinp/ipd/api"
)

func main() {
	var opts struct {
		DBPath        string `short:"f" long:"file" description:"Path to GeoIP database" value-name:"FILE" default:""`
		Listen        string `short:"l" long:"listen" description:"Listening address" value-name:"ADDR" default:":8080"`
		CORS          bool   `short:"x" long:"cors" description:"Allow requests from other domains"`
		ReverseLookup bool   `short:"r" long:"reverse-lookup" description:"Perform reverse hostname lookups"`
		PortTesting   bool   `short:"p" long:"port-testing" description:"Enable port testing"`
		Template      string `short:"t" long:"template" description:"Path to template" default:"index.html"`
	}
	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		os.Exit(1)
	}

	api := api.New()
	api.CORS = opts.CORS
	if opts.ReverseLookup {
		log.Println("Enabling reverse lookup")
		api.EnableReverseLookup()
	}
	if opts.PortTesting {
		log.Println("Enabling port testing")
		api.EnablePortTesting()
	}
	if opts.DBPath != "" {
		log.Printf("Enabling country lookup (using database: %s)\n", opts.DBPath)
		if err := api.EnableCountryLookup(opts.DBPath); err != nil {
			log.Fatal(err)
		}
	}
	api.Template = opts.Template

	log.Printf("Listening on %s", opts.Listen)
	if err := api.ListenAndServe(opts.Listen); err != nil {
		log.Fatal(err)
	}
}
