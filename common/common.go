package common

import (
    "encoding/json"
    "reflect"
    "strconv"
    "time"
)

func Number2String(number interface{}) string {
    kStr := reflect.TypeOf(number).Kind()
    switch kStr {
    case reflect.Int64:
        number = strconv.FormatInt(number.(int64), 10)
    case reflect.Int32:
        number = int64(number.(int32))
        number = strconv.FormatInt(number.(int64), 10)
    case reflect.Int:
        number = strconv.Itoa(number.(int))
    case reflect.Float64:
        number = strconv.FormatFloat(number.(float64), 'f', -1, 64)
    case reflect.Float32:
        number = float64(number.(float32))
        number = strconv.FormatFloat(number.(float64), 'f', -1, 64)
    case reflect.Bool:
        number = number.(bool)
        if number == true {
            number = "1"
        } else {
            number = "0"
        }
    case reflect.Slice, reflect.Map:
        by, _ := json.Marshal(number)
        number = string(by)
    }
    
    return number.(string)
}

func StringToTimeByFormat(timeStr, layout string) time.Time {
    loc, _ := time.LoadLocation("Asia/Shanghai")
    timeObj, _ := time.ParseInLocation(layout, timeStr, loc)
    
    return timeObj
}

func GetTomorrowZero() time.Time {
    tomorrow := time.Now().Add(24 * time.Hour)
    tomorrowZero := tomorrow.Format("2006-01-02 00:00:00")
    return StringToTimeByFormat(tomorrowZero, "2006-01-02 15:04:05")
}
