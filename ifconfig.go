package main

import (
	"encoding/json"
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/oschwald/geoip2-golang"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

var agentExp = regexp.MustCompile("^(?i)(curl|wget|fetch\\slibfetch)\\/.*$")

type Client struct {
	IP     net.IP
	JSON   string
	Header http.Header
	Cmd
}

type Cmd struct {
	Name string
	Args string
}

type Ifconfig struct {
	DB *geoip2.Reader
}

func (c *Cmd) String() string {
	return c.Name + " " + c.Args
}

func isCLI(userAgent string) bool {
	return agentExp.MatchString(userAgent)
}

func parseRealIP(req *http.Request) net.IP {
	var host string
	realIP := req.Header.Get("X-Real-IP")
	if realIP != "" {
		host = realIP
	} else {
		host, _, _ = net.SplitHostPort(req.RemoteAddr)
	}
	return net.ParseIP(host)
}

func pathToKey(path string) string {
	trimmed := strings.TrimSuffix(strings.TrimPrefix(path, "/"), ".json")
	return strings.ToLower(trimmed)
}

func isJSON(req *http.Request) bool {
	return strings.HasSuffix(req.URL.Path, ".json") ||
		strings.Contains(req.Header.Get("Accept"), "application/json")
}

func (i *Ifconfig) LookupCountry(ip net.IP) (string, error) {
	if i.DB == nil {
		return "", nil
	}
	record, err := i.DB.Country(ip)
	if err != nil {
		return "", err
	}
	country, exists := record.Country.Names["en"]
	if !exists {
		country, exists = record.RegisteredCountry.Names["en"]
		if !exists {
			return "", fmt.Errorf(
				"could not determine country for IP: %s", ip)
		}
	}
	return country, nil
}

func (i *Ifconfig) JSON(req *http.Request, key string) (string, error) {
	var header http.Header
	if key == "all" {
		header = req.Header
	} else {
		header = http.Header{
			key: []string{req.Header.Get(key)},
		}
	}
	b, err := json.MarshalIndent(header, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (i *Ifconfig) Plain(req *http.Request, key string, ip net.IP) string {
	if key == "" || key == "ip" {
		return fmt.Sprintf("%s\n", ip)
	}
	return fmt.Sprintf("%s\n", req.Header.Get(key))
}

func lookupCmd(values url.Values) Cmd {
	cmd, exists := values["cmd"]
	if !exists || len(cmd) == 0 {
		return Cmd{Name: "curl"}
	}
	switch cmd[0] {
	case "curl":
		return Cmd{Name: "curl"}
	case "fetch":
		return Cmd{Name: "fetch", Args: "-qo -"}
	case "wget":
		return Cmd{Name: "wget", Args: "-qO -"}
	}
	return Cmd{Name: "curl"}
}

func (i *Ifconfig) handler(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(w, "Invalid request method", 405)
		return
	}
	ip := parseRealIP(req)
	key := pathToKey(req.URL.Path)
	cmd := lookupCmd(req.URL.Query())
	country, err := i.LookupCountry(ip)
	if err != nil {
		log.Print(err)
	}
	req.Header["X-Ip-Country"] = []string{country}
	if isJSON(req) {
		out, err := i.JSON(req, key)
		if err != nil {
			log.Print(err)
			http.Error(w, "Failed to marshal JSON", 500)
			return
		}
		io.WriteString(w, out)
	} else if isCLI(req.UserAgent()) {
		io.WriteString(w, i.Plain(req, key, ip))
	} else {
		funcMap := template.FuncMap{
			"ToLower": strings.ToLower,
		}
		t, _ := template.
			New("index.html").
			Funcs(funcMap).
			ParseFiles("index.html")
		b, err := json.MarshalIndent(req.Header, "", "  ")
		if err != nil {
			log.Print(err)
			http.Error(w, "Failed to marshal JSON", 500)
			return
		}
		client := &Client{
			IP:     ip,
			JSON:   string(b),
			Header: req.Header,
			Cmd:    cmd,
		}
		t.Execute(w, client)
	}
}

func Create(path string) (*Ifconfig, error) {
	if path == "" {
		log.Print("Path to GeoIP database not given. Country lookup will be disabled")
		return &Ifconfig{}, nil
	}
	db, err := geoip2.Open(path)
	if err != nil {
		return nil, err
	}
	return &Ifconfig{DB: db}, nil
}

func main() {
	var opts struct {
		DBPath string `short:"f" long:"file" description:"Path to GeoIP database" value-name:"FILE" default:""`
		Listen string `short:"l" long:"listen" description:"Listening address" value-name:"ADDR" default:":8080"`
	}
	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		os.Exit(1)
	}
	i, err := Create(opts.DBPath)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", i.handler)
	log.Printf("Listening on %s", opts.Listen)
	if err := http.ListenAndServe(opts.Listen, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
