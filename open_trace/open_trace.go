package open_trace

import (
    "github.com/opentracing/opentracing-go"
    "github.com/uber/jaeger-client-go"
    jaegerCfg "github.com/uber/jaeger-client-go/config"
    "github.com/uber/jaeger-lib/metrics"
    "io"
)

// serviceName 表示 trace 名称
// traceHost   表示 trace 数据上报地址
// samplerRate 表示 trace 采样率，(0,1) 大于 1 时按照每秒进行采样
// ReportQueueSize 表示 trace 上报的并发数
// Logger 用来记录每个trace上报的日志信息
// Metrics 用于监控埋点上报
type TraceConfig struct {
    ServiceName     string
    TraceHost       string
    SamplerRate     float64
    ReportQueueSize int
    Logger          jaeger.Logger
    Metrics         metrics.Factory
}

// 初始化OpenTracing
func InitTrace(traceCfg *TraceConfig) io.Closer {
    cfg := jaegerCfg.Configuration{
        ServiceName: traceCfg.ServiceName,
    }
    var (
        sampler  jaeger.Sampler
        jLogger  jaeger.Logger
        jMetrics metrics.Factory
    )
    if traceCfg.SamplerRate <= 0 {
        panic("Trace Sampler Rate Cannot Below 0")
    }
    if traceCfg.SamplerRate < 1 && traceCfg.SamplerRate > 0 {
        sampler, _ = jaeger.NewProbabilisticSampler(traceCfg.SamplerRate)
    } else {
        sampler = jaeger.NewRateLimitingSampler(traceCfg.SamplerRate) // 定义固定每秒采样数，如每秒采样数：100
    }
    
    // 自定义 数据上报，如上报到文件日志
    report := &jaegerCfg.ReporterConfig{
        LogSpans:           true,
        LocalAgentHostPort: traceCfg.TraceHost,
    }
    _, err := report.NewReporter(traceCfg.ServiceName, nil, traceCfg.Logger)
    if err != nil {
        panic("Init OpenTracing Report Error With " + err.Error())
    }
    cfg.Reporter = report
    
    // 日志输出
    if traceCfg.Logger != nil {
        jLogger = traceCfg.Logger
    } else {
        jLogger = jaeger.StdLogger
    }
    
    // metrics 监控
    if traceCfg.Metrics != nil {
        jMetrics = metrics.NullFactory
    } else {
        jMetrics = metrics.NullFactory
    }
    
    // Initialize tracer with a logger and a metrics factory
    tracer, closer, err := cfg.NewTracer(
        jaegerCfg.Logger(jLogger),
        jaegerCfg.Metrics(jMetrics),
        jaegerCfg.Sampler(sampler),
    )
    if err != nil {
        panic("Init Trace Failed With " + err.Error())
    }
    // Set the singleton opentracing.Tracer with the Jaeger tracer.
    opentracing.SetGlobalTracer(tracer)
    
    return closer
}
