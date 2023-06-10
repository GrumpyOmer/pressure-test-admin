package logic

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"pressure-test-admin/schema"
	"pressure-test-admin/tool/go-stress-testing/model"
	"pressure-test-admin/tool/go-stress-testing/server"
	"sync"
	"time"
)

// array 自定义数组参数
type array []string

var (
	concurrency uint64 = 1       // 并发数
	totalNumber uint64 = 1       // 请求数(单个并发/协程)
	debugStr           = "false" // 是否是debug
	requestURL         = ""      // 压测的url 目前支持，http/https ws/wss
	path               = ""      // curl文件路径 http接口压测，自定义参数设置
	verify             = ""      // verify 验证方法 在server/verify中 http 支持:statusCode、json webSocket支持:json
	headers     array            // 自定义头信息传递给服务器
	body               = ""      // HTTP POST方式传送数据
	maxCon             = 1       // 单个连接最大请求数
	code               = 200     // 成功状态码
	http2              = false   // 是否开http2.0
	keepalive          = false   // 是否开启长连接
	cpuNumber          = 1       // CUP 核数，默认为一核，一般场景下单核已经够用了
	timeout     int64  = 0       // 超时时间，默认不设置
)
var upgrader = websocket.Upgrader{
	// 支持跨域
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func PressureByUrl(c *gin.Context) {
	con, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("Error during connection upgradation:" + err.Error())
		return
	}
	defer con.Close()
	go pingHandle(con)
	for {
		// road request message
		mt, message, err := con.ReadMessage()
		if err != nil {
			fmt.Println("Error road message:" + err.Error())
			break
		}
		if string(message) == "ping" {
			con.WriteMessage(mt, []byte("pong"))
			continue
		}
		currentParam := schema.PressureByUrlReq{}
		rsp := schema.PublicRsp{}
		res := []byte{}
		if err := json.Unmarshal(message, &currentParam); err != nil {
			fmt.Println("Error unmarshal message:" + err.Error())
			break
		}
		fmt.Println(currentParam)
		if currentParam.Url == "" || currentParam.ConcurrencyQuantity == 0 || currentParam.PressureTime == 0 {
			fmt.Println("Error invalid param")
			rsp.Code = 400
			rsp.Message = "Error invalid param"
		} else {
			// logic
			request, err := model.NewRequest(currentParam.Url, verify, code, 0, false, path, headers, body, maxCon, http2, keepalive)
			if err != nil {
				fmt.Println("Error invalid param")
				rsp.Code = 400
				rsp.Message = "Error invalid param"
			} else {
				fmt.Printf("\n 开始启动  并发数:%d 请求数:%d 请求参数: \n", currentParam.ConcurrencyQuantity, totalNumber)
				request.Print()
				ticker1 := time.NewTicker(time.Second)
				w := sync.WaitGroup{}
				w.Add(int(currentParam.PressureTime))
				defer ticker1.Stop()
				startTime := time.Now().UnixNano()

				for {
					// 每1秒中从chan t.C 中读取一次
					<-ticker1.C
					if currentParam.PressureTime == 0 {
						break
					}
					// 开始处理
					go server.Dispose(c, currentParam.ConcurrencyQuantity, totalNumber, request, con, &w, startTime)
					currentParam.PressureTime--
				}
				w.Wait()
				rsp.Code = 0
			}
		}
		res, _ = json.Marshal(rsp)
		err = con.WriteMessage(mt, res)
		if err != nil {
			fmt.Println("Error write message:" + err.Error())
			break
		}
	}

}

func pingHandle(c *websocket.Conn) {
	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-ticker.C:
			err := c.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				fmt.Println(err.Error())
				fmt.Println(c.Close())
				return
			}
		}
	}
}
