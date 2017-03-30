# ipd

[![Build Status](https://travis-ci.org/mpolden/ipd.svg)](https://travis-ci.org/mpolden/ipd)

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
  "ip": "127.0.0.1",
  "ip_decimal": 2130706433
}
```

## Features

* Easy to remember domain name
* Supports HTTPS
* Open source under the [BSD 3-Clause license](https://opensource.org/licenses/BSD-3-Clause)
* Fast
* Supports typical CLI tools (`curl`, `httpie`, `wget` and `fetch`)
* JSON output (optional)
* Country and city lookup through the MaxMind GeoIP database

## Why?

* To scratch an itch
* An excuse to use Go for something
* Faster than ifconfig.me

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
  -f, --country-db=FILE                                  Path to GeoIP country database
  -c, --city-db=FILE                                     Path to GeoIP city database
  -l, --listen=ADDR                                      Listening address (default: :8080)
  -r, --reverse-lookup                                   Perform reverse hostname lookups
  -p, --port-lookup                                      Enable port lookup
  -t, --template=FILE                                    Path to template (default: index.html)
  -H, --trusted-header=NAME                              Header to trust for remote IP, if present (e.g. X-Real-IP)
  -L, --log-level=[debug|info|warn|error|fatal|panic]    Log level to use (default: info)

Help Options:
  -h, --help                                             Show this help message
```
