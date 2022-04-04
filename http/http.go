package http

import (
    "encoding/json"
    "github.com/astaxie/beego/httplib"
    "github.com/reaburoa/utils/common"
    "io/ioutil"
    netHttp "net/http"
    "strings"
    "time"
)

const (
    
    HttpMethodGet   = "GET"
    HttpMethodPost  = "POST"
    
    RequestDataTypeJson = "json"
)

var (
	httpTransport netHttp.RoundTripper
    defaultUerAgent = "Default User-Agent"
)

func SetTransport(transport *netHttp.Transport) {
    httpTransport = transport
}

func SetUserAgent(ua string) {
    defaultUerAgent = ua
}

func setHeader(req *httplib.BeegoHTTPRequest, header map[string]string) *httplib.BeegoHTTPRequest {
    for key, val := range header {
        req.Header(key, val)
    }
    return req
}

func Curl(url, method string, header map[string]string, reqParams, fileList map[string]interface{}, reqType string, connTimeout, rwTimeout time.Duration) (*netHttp.Response, []byte, error) {
    req := httplib.NewBeegoRequest(url, method)
    defaultSetting := httplib.BeegoHTTPSettings{
        UserAgent:        defaultUerAgent,
        Transport:        httpTransport,
        ConnectTimeout:   connTimeout,
        ReadWriteTimeout: rwTimeout,
    }
    req.Setting(defaultSetting)
    req = setHeader(req, header)
    if strings.ToLower(reqType) == RequestDataTypeJson {
        by, _ := json.Marshal(reqParams)
        req = req.Body(by)
    } else {
        if len(reqParams) > 0 {
            for key, val := range reqParams {
                req.Param(key, common.Number2String(val))
            }
        }
    }
    if len(fileList) > 0 {
        for f, v := range fileList {
            req.PostFile(f, common.Number2String(v))
        }
    }
    resp, err := req.DoRequest()
    if err != nil {
        return nil, nil, err
    }
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, nil, err
    }
    return resp, body, nil
}
