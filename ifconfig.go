package main

import (
    "fmt"
    "html/template"
    "io"
    "log"
    "net"
    "net/http"
    "regexp"
)

type Client struct {
    IP net.IP
}

func isCli(userAgent string) bool {
    match, _ := regexp.MatchString("^(?i)(curl|wget|fetch\\slibfetch)\\/.*$",
        userAgent)
    return match
}

func handler(w http.ResponseWriter, req *http.Request) {
    if req.Method != "GET" {
        http.Error(w, "Invalid request method", 405)
        return
    }

    var host string
    var err error
    realIP := req.Header.Get("X-Real-IP")
    if realIP != "" {
        host = realIP
    } else {
        host, _, err = net.SplitHostPort(req.RemoteAddr)
    }
    ip := net.ParseIP(host)
    if err != nil {
        log.Printf("Failed to parse remote address: %s\n", req.RemoteAddr)
        http.Error(w, "Failed to parse remote address", 500)
        return
    }

    if isCli(req.UserAgent()) {
        io.WriteString(w, fmt.Sprintf("%s\n", ip))
    } else {
        t, _ := template.ParseFiles("index.html")
        client := &Client{IP: ip}
        t.Execute(w, client)
    }
}

func main() {
    http.HandleFunc("/", handler)
    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}
