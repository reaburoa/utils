package main

import (
    "fmt"
    "github.com/reaburoa/logger/logger"
)

func main() {
    fmt.Println("logger library ...")
    logger.InitLogger("sss", "./Runtime", "json", 1, 1, 2, true, true, true)
    logger.Sugar.Info("qqqq")
}
