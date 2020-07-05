## 基于zap和lumberjack封装的一个日志记录库

### 具体使用方法
在服务入口处初始化调用一次 InitLogger 进行初始化，会初始化全局变量 Sugar，后续服务只需使用 Sugar 变量即可进行日志记录。

服务运行后，每日 00:00:00 会创建新的日志文件，保证每日日志文件更新。

#### zap 和 lumberjack 库的地址：

[zap](https://github.com/uber-go/zap)

[lumberjack](https://github.com/natefinch/lumberjack)

##### 具体使用
提交配置信息后，即可轻松使用日志插件，使用全局 Sugar 变量可以完全使用zap提供的全部日志输出函数，无需关心历史文件删除、日志文件更新。

```go
    lib.InitLogger(
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
    lib.Sugar.Info("Log info")

    // 按照指定格式进行记录日志，和go的 fmt.Sprintf 格式化一致
    lib.Sugar.Infof("sds %s", "sdd")
```