package main

import (
    "context"
    "fmt"
    "github.com/opentracing/opentracing-go"
    "github.com/opentracing/opentracing-go/ext"
    "github.com/opentracing/opentracing-go/log"
    "github.com/reaburoa/utils/open_trace"
    "io/ioutil"
    "net/http"
)

func main() {
    // 初始化OpenTracing
    cfg := open_trace.TraceConfig{
        ServiceName:     "traceDemo",
        TraceHost:       "127.0.0.1:6831",
        SamplerRate:     10,
        ReportQueueSize: 1,
        //Logger:          &trace.Logger{},
    }
    closer := open_trace.InitTrace(&cfg)
    defer closer.Close()
    
    // 使用全局的Tracer
    tracer := opentracing.GlobalTracer()
    
    ctx := context.Background()
    clientSpan, ctx := opentracing.StartSpanFromContext(ctx, "clientSpan") // 通过context生成Span
    defer clientSpan.Finish()
    
    url := "http://localhost:8082/publish"
    req, _ := http.NewRequest("GET", url, nil)
    
    // Set some tags on the clientSpan to annotate that it's the client span. The additional HTTP tags are useful for debugging purposes.
    ext.SpanKindRPCClient.Set(clientSpan)
    ext.HTTPUrl.Set(clientSpan, url)
    ext.HTTPMethod.Set(clientSpan, "GET")
    
    // Inject the client span context into the headers
    spanCtx := clientSpan.Context()
    // 注入trace信息
    tracer.Inject(spanCtx, opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
    // 设置log信息，需要根据业务自行设定
    clientSpan.LogFields(
        log.String("event", "soft error"),
        log.String("type", "cache timeout"),
        log.Int("waited.millis", 1500),
    )
    clientSpan.SetTag("dddd", "ddd123")
    
    resp, _ := http.DefaultClient.Do(req)
    bb, err := ioutil.ReadAll(resp.Body)
    fmt.Println("body", string(bb), err)
    
    defer resp.Body.Close()
}
