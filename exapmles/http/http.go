package main

import (
    "context"
    "fmt"
    uHttp "github.com/reaburoa/utils/http"
    "net"
    "net/http"
    "time"
)

func main() {
    ctx := context.Background()
    url := "https://www.baidu.com"
    trans := &http.Transport{
        DialContext: (&net.Dialer{
            Timeout:   10 * time.Second,
            KeepAlive: 45 * time.Second,
        }).DialContext,
        MaxIdleConns:          10000,
        IdleConnTimeout:       60 * time.Second,
        ExpectContinueTimeout: 5 * time.Second,
    }
    client, er := uHttp.NewHttpClient(ctx, url, http.MethodGet, trans)
    if er != nil {
        fmt.Println("Http New Client Error ", er)
    }
    resp, err := client.SetUserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.141 Safari/537.36").Bytes()
    fmt.Println(string(resp), err)
    
    cc, er := uHttp.Get(url)
    if er != nil {
        fmt.Println("Http New Client Error ", er)
    }
    ret, err := cc.SetUserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.141 Safari/537.36").Bytes()
    fmt.Println(string(ret), err)
    // defer resp.Body.Close()
    // body, er := ioutil.ReadAll(resp.Body)
    // fmt.Println("body", string(body), er)
}
