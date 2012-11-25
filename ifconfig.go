package main

import (
    "fmt"
    "html/template"
    "io"
    "log"
    "net"
    "net/http"
    "regexp"
    "strings"
)

type Client struct {
    IP     net.IP
    Port   string
    Header http.Header
}

func isCli(userAgent string) bool {
    match, _ := regexp.MatchString("^(?i)(curl|wget|fetch\\slibfetch)\\/.*$",
        userAgent)
    return match
}

func parseRealIP(req *http.Request) (net.IP, string) {
    var host string
    var port string
    realIP := req.Header.Get("X-Real-IP")
    if realIP != "" {
        host = realIP
    } else {
        host, port, _ = net.SplitHostPort(req.RemoteAddr)
    }
    return net.ParseIP(host), port
}

func pathToKey(path string) string {
    re := regexp.MustCompile("^\\/|\\.json$")
    return re.ReplaceAllLiteralString(strings.ToLower(path), "")
}

func isJson(req *http.Request) bool {
    return strings.HasSuffix(req.URL.Path, ".json") ||
        strings.Contains(req.Header.Get("Accept"), "application/json")
}

func handler(w http.ResponseWriter, req *http.Request) {
    if req.Method != "GET" {
        http.Error(w, "Invalid request method", 405)
        return
    }

    ip, port := parseRealIP(req)
    header := pathToKey(req.URL.Path)

    if isCli(req.UserAgent()) {
        if header == "" || header == "ip" {
            io.WriteString(w, fmt.Sprintf("%s\n", ip))
        } else if header == "port" {
            io.WriteString(w, fmt.Sprintf("%s\n", port))
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
        client := &Client{IP: ip, Port: port, Header: req.Header}
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
