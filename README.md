# ipd

[![Build Status](https://travis-ci.org/martinp/ipd.svg)](https://travis-ci.org/martinp/ipd)

A simple service for looking up your IP address. This is the code that powers
https://ifconfig.co

## Usage

Just the business, please:

```
$ curl ifconfig.co
127.0.0.1

$ http ifconfig.co
127.0.0.1

$ wget -qO- ifconfig.co
127.0.0.1

$ fetch -qo- http://ifconfig.co
127.0.0.1
```

Country and city lookup:

```
$ http ifconfig.co/country
Elbonia

$ http ifconfig.co/city
Bornyasherk
```

As JSON:

```
$ http --json ifconfig.co
{
  "city": "Bornyasherk",
  "country": "Elbonia",
  "ip": "127.0.0.1"
}
```

Pass the appropriate flag (usually `-4` and `-6`) to your tool to switch between
IPv4 and IPv6 lookup.

The subdomains https://v4.ifconfig.co and https://v6.ifconfig.co can be used to
force IPv4 or IPv6 lookup.

## Features

* Easy to remember domain name
* Supports IPv4 and IPv6
* Supports HTTPS
* Open source under the [BSD 3-Clause license](https://opensource.org/licenses/BSD-3-Clause)
* Fast
* Supports typical CLI tools (`curl`, `httpie`, `wget` and `fetch`)
* JSON output (optional)
* Country and city lookup through the MaxMind GeoIP database

## Why?

* To scratch an itch
* An excuse to use Go for something
* Faster than ifconfig.me and has IPv6 support

## Building

Compiling requires the [Golang compiler](https://golang.org/) to be installed.
This application can be installed by using `go get`:

`go get github.com/martinp/ipd`

### Usage

```
$ ipd -h
Usage:
  ipd [OPTIONS]

Application Options:
  -f, --country-db=FILE    Path to GeoIP country database
  -c, --city-db=FILE       Path to GeoIP city database
  -l, --listen=ADDR        Listening address (default: :8080)
  -r, --reverse-lookup     Perform reverse hostname lookups
  -p, --port-lookup        Enable port lookup
  -t, --template=          Path to template (default: index.html)

Help Options:
  -h, --help               Show this help message
```
