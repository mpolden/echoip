package main

import (
    "fmt"
    "io"
    "log"
    "net"
    "net/http"
    "regexp"
)

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

    host, _, err := net.SplitHostPort(req.RemoteAddr)
    if err != nil {
        log.Printf("Failed to parse remote address: %s\n", req.RemoteAddr)
        http.Error(w, "Failed to parse remote address", 500)
        return
    }

    if isCli(req.UserAgent()) {
        io.WriteString(w, fmt.Sprintf("%s\n", host))
    } else {
        // XXX: Render HTML
    }
}

func main() {
    http.HandleFunc("/", handler)
    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}
