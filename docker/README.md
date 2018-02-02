# Ipd Docker

## Build

        docker build -t ipd .
        
## Run

        docker run -p 8080:8080 ipd 
        
        
### Run with GeoIp databases (from docker container)

        docker run -p 8080:8080 ipd -c ./GeoLite2-City.mmdb -f ./GeoLite2-Country.mmdb 
