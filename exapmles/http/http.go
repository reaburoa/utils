package main

import (
    "fmt"
    uHttp "github.com/reaburoa/utils/http"
    "net"
    "net/http"
    "time"
)

func main() {
    uHttp.SetUserAgent("test ua")
    uHttp.SetTransport(&http.Transport{
        DialContext: (&net.Dialer{
            Timeout:   10 * time.Second,
            KeepAlive: 45 * time.Second,
        }).DialContext,
        MaxIdleConns:          10000,
        IdleConnTimeout:       60 * time.Second,
        ExpectContinueTimeout: 5 * time.Second,
    })
    header := map[string]string{
        "User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.141 Safari/537.36",
    }
    resp, _, _ := uHttp.Curl("http://www.baidu.com", uHttp.HttpMethodGet, header, nil, nil, "", 1*time.Second, 5*time.Second)
    fmt.Println(resp)
}
