package main

import (
    "fmt"
    "github.com/opentracing/opentracing-go"
    "github.com/opentracing/opentracing-go/ext"
    "github.com/reaburoa/utils/open_trace"
    "log"
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
    
    http.HandleFunc("/publish", func(w http.ResponseWriter, r *http.Request) {
        // Extract the context from the headers
        tracer := opentracing.GlobalTracer() // 全局tracer
        spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header)) // 从请求中提取span
        serverSpan := tracer.StartSpan("server", ext.RPCServerOption(spanCtx)) // 开启span
        defer serverSpan.Finish()
        
        t := string([]rune(fmt.Sprint(serverSpan)[:16])) // 获取trace-id
        fmt.Println("server span ", t)
        w.Write([]byte("ok trace_id " + t))
    })
    
    log.Fatal(http.ListenAndServe(":8082", nil))
}
