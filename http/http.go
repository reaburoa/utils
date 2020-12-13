package http

import (
    "log"
    "net/http"
    "net/url"
)

type RequestHttp struct {
    url    string
    req    *http.Request
    params map[string][]string
    files  map[string]string
    resp   *http.Response
    body   []byte
}

func NewRequest(rawurl, method string) *RequestHttp {
    var resp http.Response
    u, err := url.Parse(rawurl)
    if err != nil {
        log.Println("Http:", err)
    }
    req := http.Request{
        URL:        u,
        Method:     method,
        Header:     make(http.Header),
        Proto:      "HTTP/1.1",
        ProtoMajor: 1,
        ProtoMinor: 1,
    }
    return &RequestHttp{
        url:    rawurl,
        req:    &req,
        params: map[string][]string{},
        files:  map[string]string{},
        resp:   &resp,
    }
}

func Get(url string) *RequestHttp {
    return NewRequest(url, "Get")
}

func (r *RequestHttp) byte() ([]byte, error) {
    if r.body != nil {
        return r.body, nil
    }
    return nil, nil
}