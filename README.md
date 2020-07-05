# 各类工具封装基础库，便于工程项目使用

## logger库
### 方便配置、使用的log工具类
### 使用
在服务入口处初始化调用一次 InitLogger 进行初始化，初始化全局变量 Sugar，使用 Sugar 进行日志记录。
服务运行后，每日 00:00:00 会创建新的日志文件，保证每日日志文件更新。

#### zap 和 lumberjack 库的地址：
- [zap](https://github.com/uber-go/zap)库，日志记录基础库，可以区分日志级别：debug、info、warning、error、fatal等
- [lumberjack](https://github.com/natefinch/lumberjack)库，对日志进行自动化切分、压缩、归档、删除历史日志文件

##### 使用示例
提供配置信息后，即可轻松使用日志插件，可以输出静态日志、格式化日志信息等。

```go
    logger.InitLogger(
        "log-prefix", // 服务名称，日志记录会以此参数为前缀，日志文件：log-prefix_YYYYMMDDHHIISS.log
        "./Runtime", // 日志记录位置
        "json", // 日志格式，默认使用json格式，console表示使用console \t 分割的日志风格
        1, // 日志文件最大大小，单位 MB
        10, // 日志文件最多存放多长时间，单位 天
        10, // 日志文件备份保留多少个，备份文件格式 log-prefix_20200705153229-2020-07-05T15-32-30.597.log.gz
        true, // 日志文件是否压缩
        true, // 是否开启debug
        true, // 是否输出到标准输出
    )

    // 记录静态字符串
    logger.Sugar.Info("Log info")

    // 按照指定格式进行记录日志，和go的 fmt.Sprintf 格式化一致
    logger.Sugar.Infof("sds %s", "sdd")

    logger.Sugar.Warn("Warning info ")
```