# echoip

A simple service for looking up your IP address. This is the code that powers
https://ip.level.io.

## Usage

Just the business, please:

```
$ curl ip.level.io
127.0.0.1

$ http ip.level.io
127.0.0.1

$ wget -qO- ip.level.io
127.0.0.1

$ fetch -qo- https://ip.level.io
127.0.0.1

$ bat -print=b ip.level.io/ip
127.0.0.1
```

Country and city lookup:

```
$ curl ip.level.io/country
Elbonia

$ curl ip.level.io/country-iso
EB

$ curl ip.level.io/city
Bornyasherk

$ curl ip.level.io/asn
AS59795

$ curl ip.level.io/asn-org
Hosting4Real
```

As JSON:

```
$ curl -H 'Accept: application/json' ip.level.io  # or curl ip.level.io/json
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
$ curl ip.level.io/port/80
{
  "ip": "127.0.0.1",
  "port": 80,
  "reachable": false
}
```

Pass the appropriate flag (usually `-4` and `-6`) to your client to switch
between IPv4 and IPv6 lookup.

## Features

- Easy to remember domain name
- Fast
- Supports IPv6
- Supports HTTPS
- Supports common command-line clients (e.g. `curl`, `httpie`, `ht`, `wget` and `fetch`)
- JSON output
- ASN, country and city lookup using the MaxMind GeoIP database
- Port testing
- All endpoints (except `/port`) can return information about a custom IP address specified via `?ip=` query parameter
- Open source under the [BSD 3-Clause license](https://opensource.org/licenses/BSD-3-Clause)

## Why?

- To scratch an itch
- An excuse to use Go for something
- Faster than ifconfig.me and has IPv6 support

### Usage

```
$ echoip -h
Usage of echoip:
  -C int
    	Size of response cache. Set to 0 to disable
  -H value
    	Header to trust for remote IP, if present (e.g. X-Real-IP)
  -P	Enables profiling handlers
  -S string
    	IP Stack API Key
  -a string
    	Path to GeoIP ASN database
  -c string
    	Path to GeoIP city database
  -d string
    	Which database to use, 'ipstack' or 'geoip' (default "geoip")
  -f string
    	Path to GeoIP country database
  -h	Use HTTPS for IP Stack ( only non-free accounts )
  -l string
    	Listening address (default ":8080")
  -p	Enable port lookup
  -r	Perform reverse hostname lookups
  -s	Show sponsor logo
  -t string
    	Path to template dir (default "html")
  -x	Enable security module for IP Stack ( must have security module, aka. non-free account. )
```
