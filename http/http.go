package http

import (
    "fmt"
    "github.com/astaxie/beego/httplib"
    "github.com/reaburoa/utils/common"
    "net/http"
    "strings"
    "time"
)

type Servicer interface {
    GetSrvName() string
    GetHost() string
    Header() map[string]string
}

const (
    defaultUerAgent = "Default User-Agent"
    HttpMethodGet   = "GET"
    HttpMethodPost  = "POST"
)

var httpTransport http.RoundTripper

func SetTransport(transport *http.Transport) {
    httpTransport = transport
}

func setHeader(req *httplib.BeegoHTTPRequest, srv Servicer) *httplib.BeegoHTTPRequest {
    for key, val := range srv.Header() {
        req.Header(key, val)
    }
    return req
}

func getReqUrl(srv Servicer, uri string) string {
    if strings.Index(uri, "http://") >= 0 || strings.Index(uri, "https://") >= 0 {
        return uri
    }
    return fmt.Sprintf("%s%s", srv.GetHost(), uri)
}

func Curl(srv Servicer, uri, method string, reqParams map[string]interface{}, connTimeout, rwTimeout time.Duration) ([]byte, error) {
    url := getReqUrl(srv, uri)
    req := httplib.NewBeegoRequest(url, method)
    defaultSetting := httplib.BeegoHTTPSettings{
        UserAgent:        defaultUerAgent,
        Transport:        httpTransport,
        ConnectTimeout:   connTimeout,
        ReadWriteTimeout: rwTimeout,
    }
    req.Setting(defaultSetting)
    req = setHeader(req, srv)
    if len(reqParams) > 0 {
        for key, val := range reqParams {
            req.Param(key, common.Number2String(val))
        }
    }
    resp, err := req.Bytes()
    if err != nil {
        return nil, err
    }
    return resp, nil
}