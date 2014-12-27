FROM golang:onbuild

EXPOSE 8080

ADD http://geolite.maxmind.com/download/geoip/database/GeoLite2-Country.mmdb.gz /var/tmp/
ADD index.html /go/bin/
RUN gunzip /var/tmp/GeoLite2-Country.mmdb.gz && \
    chown nobody:nogroup /var/tmp/GeoLite2-Country.mmdb
USER nobody

CMD ["-f", "/var/tmp/GeoLite2-Country.mmdb"]
ENTRYPOINT ["/go/bin/app"]
