package http

import (
    "bytes"
    "compress/gzip"
    "context"
    "encoding/json"
    "encoding/xml"
    "fmt"
    "io"
    "io/ioutil"
    "log"
    "mime/multipart"
    "net/http"
    "net/http/cookiejar"
    "net/textproto"
    "net/url"
    "os"
    "path"
    "path/filepath"
    "strings"
    "sync"
    "time"
)

var (
    defaultCookieJar http.CookieJar
    settingMutex     sync.Mutex
)

func Get(url string) (*HttpClient, error) {
    return NewHttpClient(context.Background(), url, http.MethodGet, nil)
}

func Post(url string, params map[string]string) (*HttpClient, error) {
    client, err := NewHttpClient(context.Background(), url, http.MethodPost, nil)
    if err != nil {
        return nil, err
    }
    return client.MultiParams(params), nil
}

// createDefaultCookie creates a global cookiejar to store cookies.
func createDefaultCookie() {
    settingMutex.Lock()
    defer settingMutex.Unlock()
    defaultCookieJar, _ = cookiejar.New(nil)
}

type HttpClient struct {
    url             string
    files           map[string]string // 上传文件form表单名以及文件路径
    fileContentType string
    request         *http.Request
    client          *http.Client
    resp            *http.Response
    params          map[string][]string
    userAgent       string
    retry           int
    retryDelay      time.Duration
    enableCookie    bool
    body            []byte
    gzip            bool
}

func NewHttpClient(ctx context.Context, urlPath, method string, trans http.RoundTripper) (*HttpClient, error) {
    if trans == nil {
        trans = http.DefaultTransport
    }
    c := &http.Client{
        Transport: trans,
    }
    u, er := url.Parse(urlPath)
    if er != nil {
        return nil, er
    }
    req := &http.Request{
        URL:        u,
        Method:     method,
        Header:     make(http.Header),
        Proto:      "HTTP/1.1",
        ProtoMajor: 1,
        ProtoMinor: 1,
    }
    return &HttpClient{
        client:  c,
        request: req,
        files:   make(map[string]string),
        resp:    &http.Response{},
        params:  map[string][]string{},
        body:    []byte{},
        url:     urlPath,
    }, nil
}

func (h *HttpClient) SetGzipOn(bl bool) *HttpClient {
    h.gzip = bl
    
    return h
}

func (h *HttpClient) SetCookie(jar http.CookieJar) *HttpClient {
    h.client.Jar = jar
    return h
}

func (h *HttpClient) SetUserAgent(ua string) *HttpClient {
    h.userAgent = ua
    
    return h
}

func (h *HttpClient) SetRetries(retry int, retryDelay time.Duration) *HttpClient {
    h.retry = retry
    h.retryDelay = retryDelay
    
    return h
}

func (h *HttpClient) getResponse() (*http.Response, error) {
    if h.resp.StatusCode != 0 {
        return h.resp, nil
    }
    resp, err := h.DoRequest()
    if err != nil {
        return nil, err
    }
    h.resp = resp
    return resp, nil
}

func (h *HttpClient) Header(key, value string) *HttpClient {
    h.request.Header.Set(key, value)
    return h
}

func (h *HttpClient) MultiHeader(header map[string]string) *HttpClient {
    if len(header) <= 0 {
        return h
    }
    for key, value := range header {
        h.request.Header.Set(key, value)
    }
    return h
}

func (h *HttpClient) Param(key, value string) *HttpClient {
    if vl, ok := h.params[key]; ok {
        h.params[key] = append(vl, value)
    } else {
        h.params[key] = []string{value}
    }
    
    return h
}

func (h *HttpClient) MultiParams(params map[string]string) *HttpClient {
    for key, value := range params {
        h.Param(key, value)
    }
    
    return h
}

func (h *HttpClient) PostFile(formName, filename string) *HttpClient {
    h.files[formName] = filename
    
    return h
}

func (h *HttpClient) SetFileContentType(contentType string) *HttpClient {
    h.fileContentType = contentType
    
    return h
}

func (h *HttpClient) fileWriter(fieldName, fileName string) (io.Writer, error) {
    body := new(bytes.Buffer)
    writer := multipart.NewWriter(body)
    m := make(textproto.MIMEHeader)
    m.Set("Content-Disposition",
        fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
            fieldName, // 参数名为file
            filepath.Base(fileName)))
    // 设置文件格式
    if h.fileContentType == "" { // 文件的数据格式默认为数据流
        m.Set("Content-Type", "application/octet-stream")
    } else {
        m.Set("Content-Type", h.fileContentType)
    }
    return writer.CreatePart(m)
}

func (h *HttpClient) buildURL(paramBody string) {
    // build GET url with query string
    if h.request.Method == "GET" && len(paramBody) > 0 {
        if strings.Contains(h.url, "?") {
            h.url += "&" + paramBody
        } else {
            h.url = h.url + "?" + paramBody
        }
        return
    }
    
    // build POST/PUT/PATCH url and body
    if (h.request.Method == "POST" || h.request.Method == "PUT" || h.request.Method == "PATCH" || h.request.Method == "DELETE") && h.request.Body == nil {
        // with files
        if len(h.files) > 0 {
            pr, pw := io.Pipe()
            bodyWriter := multipart.NewWriter(pw)
            go func() {
                for formname, filename := range h.files {
                    fileWriter, err := h.fileWriter(formname, filename)
                    if err != nil {
                        log.Println("Httplib:", err)
                    }
                    fh, err := os.Open(filename)
                    if err != nil {
                        log.Println("Httplib:", err)
                    }
                    // iocopy
                    _, err = io.Copy(fileWriter, fh)
                    fh.Close()
                    if err != nil {
                        log.Println("Httplib:", err)
                    }
                }
                for k, v := range h.params {
                    for _, vv := range v {
                        bodyWriter.WriteField(k, vv)
                    }
                }
                bodyWriter.Close()
                pw.Close()
            }()
            h.Header("Content-Type", bodyWriter.FormDataContentType())
            h.request.Body = ioutil.NopCloser(pr)
            h.Header("Transfer-Encoding", "chunked")
            return
        }
        
        // with params
        if len(paramBody) > 0 {
            h.Header("Content-Type", "application/x-www-form-urlencoded")
            h.Body(paramBody)
        }
    }
}

func (h *HttpClient) Body(data interface{}) *HttpClient {
    switch t := data.(type) {
    case string:
        bf := bytes.NewBufferString(t)
        h.request.Body = ioutil.NopCloser(bf)
        h.request.ContentLength = int64(len(t))
    case []byte:
        bf := bytes.NewBuffer(t)
        h.request.Body = ioutil.NopCloser(bf)
        h.request.ContentLength = int64(len(t))
    }
    return h
}

// XMLBody adds request raw body encoding by XML.
func (h *HttpClient) XMLBody(obj interface{}) (*HttpClient, error) {
    if h.request.Body == nil && obj != nil {
        byts, err := xml.Marshal(obj)
        if err != nil {
            return h, err
        }
        h.request.Body = ioutil.NopCloser(bytes.NewReader(byts))
        h.request.ContentLength = int64(len(byts))
        h.request.Header.Set("Content-Type", "application/xml")
    }
    
    return h, nil
}

func (h *HttpClient) JsonBody(obj interface{}) (*HttpClient, error) {
    if h.request.Body == nil && obj != nil {
        by, err := json.Marshal(obj)
        if err != nil {
            return h, err
        }
        h.request.Body = ioutil.NopCloser(bytes.NewReader(by))
        h.request.ContentLength = int64(len(by))
        h.request.Header.Set("Content-Type", "application/json")
    }
    return h, nil
}

func (h *HttpClient) DoRequest() (resp *http.Response, err error) {
    var paramBody string
    if len(h.params) > 0 {
        var buf bytes.Buffer
        for k, v := range h.params {
            for _, vv := range v {
                buf.WriteString(url.QueryEscape(k))
                buf.WriteByte('=')
                buf.WriteString(url.QueryEscape(vv))
                buf.WriteByte('&')
            }
        }
        paramBody = buf.String()
        paramBody = paramBody[0 : len(paramBody)-1]
    }
    
    h.buildURL(paramBody)
    u, er := url.Parse(h.url)
    if er != nil {
        return nil, er
    }
    h.request.URL = u
    if h.enableCookie {
        if defaultCookieJar == nil {
            createDefaultCookie()
        }
        h.client.Jar = defaultCookieJar
    }
    if h.userAgent != "" && h.request.Header.Get("User-Agent") == "" {
        h.Header("User-Agent", h.userAgent)
    }
    
    // retries default value is 0, it will run once.
    // retries equal to -1, it will run forever until success
    // retries is setted, it will retries fixed times.
    // Sleeps for a 400ms in between calls to reduce spam
    for i := 0; h.retry == -1 || i <= h.retry; i++ {
        resp, err = h.client.Do(h.request)
        if err == nil {
            break
        }
        time.Sleep(h.retryDelay)
    }
    return resp, err
}

func (h *HttpClient) Response() (*http.Response, error) {
    return h.getResponse()
}

// ToJSON returns the map that marshals from the body bytes as json in response .
// it calls Response inner.
func (h *HttpClient) ToJSON(v interface{}) error {
    data, err := h.Bytes()
    if err != nil {
        return err
    }
    return json.Unmarshal(data, v)
}

// ToXML returns the map that marshals from the body bytes as xml in response .
// it calls Response inner.
func (h *HttpClient) ToXML(v interface{}) error {
    data, err := h.Bytes()
    if err != nil {
        return err
    }
    return xml.Unmarshal(data, v)
}

// Bytes returns the body []byte in response.
// it calls Response inner.
func (h *HttpClient) Bytes() ([]byte, error) {
    if len(h.body) > 0 {
        return h.body, nil
    }
    resp, err := h.getResponse()
    if err != nil {
        return nil, err
    }
    if resp.Body == nil {
        return nil, nil
    }
    defer resp.Body.Close()
    if h.gzip && resp.Header.Get("Content-Encoding") == "gzip" {
        reader, err := gzip.NewReader(resp.Body)
        if err != nil {
            return nil, err
        }
        h.body, err = ioutil.ReadAll(reader)
        return h.body, err
    }
    h.body, err = ioutil.ReadAll(resp.Body)
    return h.body, err
}

// ToFile saves the body data in response to one file.
// it calls Response inner.
func (h *HttpClient) ToFile(filename string) error {
    resp, err := h.getResponse()
    if err != nil {
        return err
    }
    if resp.Body == nil {
        return nil
    }
    defer resp.Body.Close()
    err = pathExistAndMkdir(filename)
    if err != nil {
        return err
    }
    f, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer f.Close()
    _, err = io.Copy(f, resp.Body)
    return err
}

// Check that the file directory exists, there is no automatically created
func pathExistAndMkdir(filename string) (err error) {
    filename = path.Dir(filename)
    _, err = os.Stat(filename)
    if err == nil {
        return nil
    }
    if os.IsNotExist(err) {
        err = os.MkdirAll(filename, os.ModePerm)
        if err == nil {
            return nil
        }
    }
    return err
}
