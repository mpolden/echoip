# echoip

[![Build Status](https://travis-ci.org/mpolden/echoip.svg)](https://travis-ci.org/mpolden/echoip)

A simple service for looking up your IP address. This is the code that powers
https://ifconfig.co.

## Usage

Just the business, please:

```
$ curl ifconfig.co
127.0.0.1

$ http ifconfig.co
127.0.0.1

$ wget -qO- ifconfig.co
127.0.0.1

$ fetch -qo- https://ifconfig.co
127.0.0.1

$ bat -print=b ifconfig.co/ip
127.0.0.1
```

Country and city lookup:

```
$ curl ifconfig.co/country
Elbonia

$ curl ifconfig.co/country-iso
EB

$ curl ifconfig.co/city
Bornyasherk

$ curl ifconfig.co/asn
AS59795
```

As JSON:

```
$ curl -H 'Accept: application/json' ifconfig.co  # or curl ifconfig.co/json
{
  "city": "Bornyasherk",
  "country": "Elbonia",
  "country_iso": "EB",
  "ip": "127.0.0.1",
  "ip_decimal": 2130706433,
  "asn": "AS59795",
  "asn_org": "Hosting4Real"
}
```

Port testing:

```
$ curl ifconfig.co/port/80
{
  "ip": "127.0.0.1",
  "port": 80,
  "reachable": false
}
```

Pass the appropriate flag (usually `-4` and `-6`) to your client to switch
between IPv4 and IPv6 lookup.

## Features

* Easy to remember domain name
* Fast
* Supports IPv6
* Supports HTTPS
* Supports common command-line clients (e.g. `curl`, `httpie`, `wget` and `fetch`)
* JSON output
* ASN, country and city lookup using the MaxMind GeoIP database
* Port testing
* Open source under the [BSD 3-Clause license](https://opensource.org/licenses/BSD-3-Clause)

## Why?

* To scratch an itch
* An excuse to use Go for something
* Faster than ifconfig.me and has IPv6 support

## Building

Compiling requires the [Golang compiler](https://golang.org/) to be installed.
This package can be installed with `go get`:

`go get github.com/mpolden/echoip/...`

For more information on building a Go project, see the [official Go
documentation](https://golang.org/doc/code.html).

## Docker image

A Docker image is available on [Docker
Hub](https://hub.docker.com/r/mpolden/echoip), which can be downloaded with:

`docker pull mpolden/echoip`

### Usage

```
$ echoip -h
Usage:
  echoip [OPTIONS]

Application Options:
  -f, --country-db=FILE        Path to GeoIP country database
  -c, --city-db=FILE           Path to GeoIP city database
  -a, --asn-db=FILE            Path to GeoIP ASN database
  -l, --listen=ADDR            Listening address (default: :8080)
  -r, --reverse-lookup         Perform reverse hostname lookups
  -p, --port-lookup            Enable port lookup
  -t, --template=FILE          Path to template (default: index.html)
  -H, --trusted-header=NAME    Header to trust for remote IP, if present (e.g. X-Real-IP)

Help Options:
  -h, --help                   Show this help message
```
