package logger

import (
    "fmt"
    "github.com/reaburoa/utils/common"
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
    "gopkg.in/natefinch/lumberjack.v2"
    "os"
    "strings"
    "sync"
    "time"
)

var (
    Sugar    *zap.SugaredLogger
    filename string
    syncOnce = sync.Once{}
)

// 使用 lumberjack 库设置log归档、切分
func getLumberJackLogger(maxSize, dayExpire, backupExpire int, compress bool) *lumberjack.Logger {
    return &lumberjack.Logger{
        Filename:   filename,
        MaxSize:    maxSize,
        MaxAge:     dayExpire,
        MaxBackups: backupExpire,
        Compress:   compress,
        LocalTime:  true,
    }
}

// 初始化日志对象
// serviceName string 服务名称
// path string 日志记录位置
// logStyle string 日志输出格式，使用json 或者 console的空格分割，默认使用 json格式
// maxSize int 文件最大大小，达到后会自动切分文件，单位：MB
// dayExpire int 日志文件留存多少时间，单位：天， 0 表示不会按照时间轮转删除旧日志文件
// backupExpire int 日志文件最多备份多少个， 0 表示全部保留
// compress bool 日志文件是否压缩
// debug bool 是否开启debug
// stdout bool 是否输出到标准输出
func InitLogger(serviceName, path, logStyle string, maxSize, dayExpire, backupExpire int, compress, debug, stdout bool) {
    setFilename(path, serviceName)
    printOut("Init And Logger File %s ...", filename)
    lumberJackLogger := getLumberJackLogger(maxSize, dayExpire, backupExpire, compress)
    encoder := getEncoder()
    var code zapcore.Encoder
    if logStyle == "" {
        logStyle = "json"
    }
    if strings.ToLower(logStyle) == "json" {
        code = zapcore.NewJSONEncoder(encoder)
    } else {
        code = zapcore.NewConsoleEncoder(encoder)
    }
    
    atomicLevel := zap.NewAtomicLevel()
    if debug {
        atomicLevel.SetLevel(zap.DebugLevel)
    } else {
        atomicLevel.SetLevel(zap.InfoLevel)
    }
    
    core := zapcore.NewCore(
        code,                                // 设置编码器
        getWriter(stdout, lumberJackLogger), // 设置日志打印方式
        atomicLevel,                         // 日志级别
    )
    
    caller := zap.AddCaller()                               // 开启开发模式，堆栈跟踪
    development := zap.Development()                        // 开启文件及行号
    filed := zap.Fields(zap.String("service", serviceName)) // 设置初始化字段
    logger := zap.New(core, caller, development, filed)
    Sugar = logger.Sugar()
    
    printOut("Logger Init Ok ...")
    if stdout == false {
        syncOnce.Do(func() {
            go update(serviceName, path, logStyle, maxSize, dayExpire, backupExpire, compress, debug, stdout)
        })
    }
}

// 获取日志写入目标
func getWriter(stdout bool, logger *lumberjack.Logger) zapcore.WriteSyncer {
    if stdout {
        return zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(logger))
    } else {
        return zapcore.NewMultiWriteSyncer(zapcore.AddSync(logger))
    }
}

// 获取日志编码格式
func getEncoder() zapcore.EncoderConfig {
    return zapcore.EncoderConfig{
        TimeKey:        "time",
        LevelKey:       "level",
        NameKey:        "logger",
        CallerKey:      "line",
        MessageKey:     "msg",
        StacktraceKey:  "stacktrace",
        LineEnding:     zapcore.DefaultLineEnding,
        EncodeLevel:    zapcore.LowercaseLevelEncoder,  // 小写编码器
        EncodeTime:     zapcore.ISO8601TimeEncoder,     // ISO8601 UTC 时间格式
        EncodeDuration: zapcore.SecondsDurationEncoder, //
        EncodeCaller:   zapcore.FullCallerEncoder,      // 全路径编码器
        EncodeName:     zapcore.FullNameEncoder,
    }
}

// 异步更新日志记录文件，每天创建一个新的日志文件
func update(serviceName, path, logStyle string, maxSize, dayExpire, backupExpire int, compress, debug, stdout bool) {
    printOut("Sync Slice Log ...")
    for {
        printOut("Start Sync Slice Log Goroutine...")
        now := time.Now()
        tomorrowTime := common.GetTomorrowZero()
        t := time.NewTimer(tomorrowTime.Sub(now))
        select {
        case <-t.C:
            printOut("Update Logger Info")
            InitLogger(serviceName, path, logStyle, maxSize, dayExpire, backupExpire, compress, debug, stdout)
        }
    }
}

// 设置日志文件
func setFilename(path, service string) {
    format := time.Now().Format("20060102")
    if path[len(path)-1:] == "/" {
        filename = fmt.Sprintf("%s%s_%s.log", path, service, format)
    } else {
        filename = fmt.Sprintf("%s/%s_%s.log", path, service, format)
    }
}

// 向stdout中输出日志信息
func printOut(format string, param ...interface{}) {
    fmt.Print(fmt.Sprintf("%s ", time.Now().Format("2006/01/02 15:04:05")))
    fmt.Println(fmt.Sprintf(format, param...))
}
