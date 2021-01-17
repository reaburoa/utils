package common

import (
    "encoding/json"
    "reflect"
    "strconv"
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
