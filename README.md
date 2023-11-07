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
- Supports IP Stack API or GeoIP
- JWT Authentication

### Installation from Release

- Download release file.
- Install `./echoip` binary ( `sudo install echoip /usr/local/bin/echoip` )
- Install configuration file( `sudo install -D etc/echoip/config.toml /etc/echoip/config.toml` )
- Point `config.TemplateDir` to release `html/`

### Installation from Source

- Install Go 1.18
- `$ cd echoip/`
- `$ make install`

### Usage

```
$ echoip
```

### Configuration

Configuration is managed in the `etc/echoip/config.toml` file. This file should be located in the `/etc` folder on your server ( /etc/echoip/config.toml ). If you have the project on your server, you can run `make install-config` to copy it the right location.

***Change location of config file with the `echoip -c /path/to/config.toml` flag***

```toml
Listen = ":8080"
TemplateDir = "html" # The directory of the template files ( eg, index.html )
RedisUrl = "redis://localhost:6379" # Redis Connection URL, leave blank for no Cache
CacheTtl = 3600 # in seconds
ReverseLookup = true
PortLookup = true
ShowSponsor = true
Database = "ipstack" # use "IP Stack" or "GeoIP"
TrustedHeaders = [] # Which header to trust, eg, `["X-Real-IP"]`
Profile = false # enable debug / profiling

[Jwt]
Enabled = false
Secret = ""

[IPStack]
ApiKey = "" 
UseHttps = true
EnableSecurity = true

[GeoIP]
CountryFile = ""
CityFile = ""
AsnFile = ""
```

### Environment Variables for Configuration
You can also use environment variables for configuration, most likely used for Docker. Configuration file takes precedence first, and then environment variables. Remove the value from the config file if you wish to use the environment variable.

```
ECHOIP_LISTEN=":8080"
ECHOIP_TEMPLATE_DIR="html/"
ECHOIP_REDIS_URL="redis://localhost:6379"
ECHOIP_DATABASE="ipstack"
ECHOIP_TRUSTED_HEADERS="X-Real-IP,X-Forwaded-For"
ECHOIP_IPSTACK_API_KEY="askdfj39sjdkf29dsjfk39sdfkj3"
ECHOIP_GEOIP_COUNTRY_FILE="/full/path/to/file.db"
ECHOIP_GEOIP_CITY_FILE="/full/path/to/file.db"
ECHOIP_GEOIP_ASN_FILE="/full/path/to/file.db"
ECHOIP_CACHE_TTL=3600
ECHOIP_REVERSE_LOOKUP=true
ECHOIP_PORT_LOOKUP=true
ECHOIP_SHOW_SPONSOR=true
ECHOIP_PROFILE=false
ECHOIP_IPSTACK_USE_HTTPS=true
ECHOIP_IPSTACK_ENABLE_SECURITY=true
ECHOIP_JWT_AUTH=false
ECHOIP_JWT_SIGNING_METHOD=HS256
ECHOIP_JWT_SECRET="HS256"
```

### Authenticate each API request with JWT

You can authenticate each API request with JWT token.
Just enable `config.Jwt.Enabled` and add your JWT secret to `config.Jwt.Secret`. 

EchoIP validates JWT signing algorithm, `config.SigningMethod` should be one of available from `golang-jwt/jwt` and match your expceted algorithm.
`config.SigningMethod string`

```
# ES256 | ES384 | ES512 
# RS256 | RS384 | RS512 
# HS256 | HS384 | HS512
```

Requests will be accepted if a valid token is provided in `Authorization: Bearer $token` header.

A `401` will be returned should the token not be valid.

***You can convert a key created with ssh-keygen using something like `ssh-keygen -f id_rsa.pub -e -mpem`***

### Caching with Redis

You can connect EchoIP to a Redis client to cache each request per IP. You can configure the life of the key in `config.CacheTtl`.

### Running with `systemd`

There is a systemd service file you can install in `/etc/systemd`.
