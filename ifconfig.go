package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"regexp"
	"strings"
)

var agentExp = regexp.MustCompile("^(?i)(curl|wget|fetch\\slibfetch)\\/.*$")

type Client struct {
	IP     net.IP
	JSON   string
	Header http.Header
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
	re := regexp.MustCompile("^\\/|\\.json$")
	return re.ReplaceAllLiteralString(strings.ToLower(path), "")
}

func isJSON(req *http.Request) bool {
	return strings.HasSuffix(req.URL.Path, ".json") ||
		strings.Contains(req.Header.Get("Accept"), "application/json")
}

func handler(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(w, "Invalid request method", 405)
		return
	}

	ip := parseRealIP(req)
	header := pathToKey(req.URL.Path)

	if isJSON(req) {
		if header == "all" {
			b, _ := json.MarshalIndent(req.Header, "", "  ")
			io.WriteString(w, fmt.Sprintf("%s\n", b))
		} else {
			m := map[string]string{
				header: req.Header.Get(header),
			}
			b, _ := json.MarshalIndent(m, "", "  ")
			io.WriteString(w, fmt.Sprintf("%s\n", b))
		}
	} else if isCLI(req.UserAgent()) {
		if header == "" || header == "ip" {
			io.WriteString(w, fmt.Sprintf("%s\n", ip))
		} else {
			value := req.Header.Get(header)
			io.WriteString(w, fmt.Sprintf("%s\n", value))
		}
	} else {
		funcMap := template.FuncMap{
			"ToLower": strings.ToLower,
		}
		t, _ := template.
			New("index.html").
			Funcs(funcMap).
			ParseFiles("index.html")
		b, _ := json.MarshalIndent(req.Header, "", "  ")
		client := &Client{IP: ip, JSON: string(b), Header: req.Header}
		t.Execute(w, client)
	}
}

func main() {
	http.Handle("/assets/", http.StripPrefix("/assets/",
		http.FileServer(http.Dir("assets/"))))
	http.HandleFunc("/", handler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
