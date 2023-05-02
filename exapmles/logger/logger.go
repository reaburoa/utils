package main

import "github.com/reaburoa/utils/logger"

func main() {
    logger.InitLogger(
        "log-prefix", // 服务名称，日志记录会以此参数为前缀，日志文件：log-prefix_YYYYMMDD.log
        "./Runtime",  // 日志记录位置
        "json",       // 日志格式，默认使用json格式，console表示使用console \t 分割的日志风格
        1,            // 日志文件最大大小，单位 MB
        10,           // 日志文件最多存放多长时间，单位 天
        10,           // 日志文件备份保留多少个，备份文件格式 log-prefix_20200705153229-2020-07-05T15-32-30.597.log.gz
        true,         // 日志文件是否压缩
        true,         // 是否开启debug
        false,        // 是否输出到标准输出
    )
    for {
        // 记录静态字符串
        logger.Sugar.Info("Log info")
        
        // 按照指定格式进行记录日志，和go的 fmt.Sprintf 格式化一致
        logger.Sugar.Infof("sds %s", "sdd")
        
        logger.Sugar.Warn("Warning info ")
    }
}
