package main

import (
    "fmt"
    "github.com/reaburoa/logger/lib"
)

func main() {
    fmt.Println("logger library ...")
    
    lib.InitLogger("newService", "/Users/zhanglei/GoApplications/logger", 1, 10, 10, true, true, true)
    
    for {
        lib.Sugar.Info(`## ReadBook Job
### Crontab job任务
- 每天12:10 启动生产者，查询昨日读书同步信息
- 每天12:20 启动消费者，消费读书同步信息

#### 生产者产生信息流程
1. 从梵高处查询信息是否同步，今日查询昨日的readBook数据，接口[WIKI](https://km.qutoutiao.net/pages/viewpage.action?pageId=64428374) 【获取 'T+1' 数据同步完整性标记】
2. 向qe提交请求，接口[WIKI](https://km.qutoutiao.net/pages/viewpage.action?pageId=91469958)【/api/job/submit】获取任务ID
3. 从过步骤2的任务ID查询任务结果，接口[WIKI](https://km.qutoutiao.net/pages/viewpage.action?pageId=91469958)【/api/job/status/】最多轮询查询100次
4. 步骤3返回成功后，查询文件的链接地址，接口[WIKI](https://km.qutoutiao.net/pages/viewpage.action?pageId=91469958)【/api/job/result/】，将文件下载（shell脚本：/data/biz/qttPush/./push.sh）到本地文件中：/data/biz/readBook/Ymd/readBookDetail.csv
5. goroutine协程，读取文件内容，循环将文件每一行数据放入channel - fileRowsChan 中
6. goroutine协程，梵高5个协程消费 channel - fileRowsChan，gid：YYYMMDD+bookID，查询近90天内的读取这本书的活跃用户，梵高接口[WIKI](https://km.qutoutiao.net/pages/viewpage.action?pageId=64428374)【/user_profile/v1/group_count_up】圈人，将人的信息上传到oss中
7. 将书的信息中添加已圈人的文件oss地址，并将书的信息放入channel - pushChannel中
8. 消费channel - pushChannel， 将以上步骤获取到的数据推入到nsq中，如果链接nsq失败，则会将channel数据消费出来，同步统计信息到prometheus

#### 消费者消费信息流程
1. 建立和nsq的链接，如果链接失败，则发送微信报警消息
2. 通过book - hashid去查询书的信息，接口[WIKI](https://km.qutoutiao.net/pages/viewpage.action?pageId=155927497)【/book/getManyBookInfos】
3. 判断书是否下架，下架的书无法推送
4. 从已经设置的文案中随机获取一个文案作为推送的title标题，并设置推送跳转url
5. 调用中台的file推送方式将数据推送出去

## 个推脚本任务
启动两个goroutine，分别做以下两个事情：

1. 监控个推的数据量，从redis的10个list中查询总数，将数据同步到prometheus
2. 启动10个goroutine消费redis的10个队列，每个goroutine消费一个redis list

### 个推流程
1. 从redis的list中获取数据，并decode到pushInfo的struct中
2. 从用户画像中获取用户的guid，接口[WIKI](http://yapi.qutoutiao.net/project/601/interface/api/31367)【/v2/userInfo】
3. 组织个推数据内容，调用个推接口，推送出去

同时启动20个goroutine进行消费处理`)
        lib.Sugar.Error("Error info...")
        lib.Sugar.Debug("dddddd")
        // lib.SugarLogger.Panic("panic err")
    
    
        //time.Sleep(5 * time.Second)
    }
    
}
