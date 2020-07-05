package lib

import (
    "fmt"
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
    "gopkg.in/natefinch/lumberjack.v2"
    "os"
    "time"
)

var (
    Sugar    *zap.SugaredLogger
    filename string
)

// 使用 lumberjack 库设置log归档、切分
func getLumberJackLogger(filename string, maxSize, dayExpire, backupExpire int, compress bool) *lumberjack.Logger {
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
// maxSize int 文件最大大小，达到后会自动切分文件，单位：MB
// dayExpire int 日志文件留存多少时间，单位：天
// backupExpire int 日志文件最多备份多少个
// compress bool 日志文件是否压缩
// debug bool 是否开启debug
// stdout bool 是否输出到标准输出
func InitLogger(serviceName, path string, maxSize, dayExpire, backupExpire int, compress, debug, stdout bool) {
    filename = fmt.Sprintf("%s/%s_%s.log", path, serviceName, time.Now().Format("20060102150405"))
    lumberJackLogger := getLumberJackLogger(filename, maxSize, dayExpire, backupExpire, compress)
    encoder := getEncoder()
    
    atomicLevel := zap.NewAtomicLevel()
    if debug {
        atomicLevel.SetLevel(zap.DebugLevel)
    } else {
        atomicLevel.SetLevel(zap.InfoLevel)
    }
    
    core := zapcore.NewCore(
        zapcore.NewJSONEncoder(encoder),     // 设置编码器
        getWriter(stdout, lumberJackLogger), // 设置日志打印方式
        atomicLevel,                         // 日志级别
    )
    
    caller := zap.AddCaller()                               // 开启开发模式，堆栈跟踪
    development := zap.Development()                        // 开启文件及行号
    filed := zap.Fields(zap.String("service", serviceName)) // 设置初始化字段
    logger := zap.New(core, caller, development, filed)
    Sugar = logger.Sugar()
    
    go update(serviceName, path, maxSize, dayExpire, backupExpire, compress, debug, stdout)
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
func update(serviceName, path string, maxSize, dayExpire, backupExpire int, compress, debug, stdout bool) {
    now := time.Now()
    tomorrowTime := time.Now().Add(24 * time.Hour)
    tomorrowZeroTime := time.Date(tomorrowTime.Year(), tomorrowTime.Month(), tomorrowTime.Day(), 0, 0, 0, 0, tomorrowTime.Location())
    t := time.NewTimer(tomorrowZeroTime.Sub(now))
    select {
    case <-t.C:
        InitLogger(serviceName, path, maxSize, dayExpire, backupExpire, compress, debug, stdout)
    }
}
