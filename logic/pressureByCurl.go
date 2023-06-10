package logic

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"pressure-test-admin/schema"
	"pressure-test-admin/tool/go-stress-testing/model"
	"pressure-test-admin/tool/go-stress-testing/server"
	"sync"
	"time"
)

func PressureByCurl(c *gin.Context) {
	con, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("Error during connection upgradation:" + err.Error())
		return
	}
	defer con.Close()
	go pingHandle(con)
	currentParam := schema.PressureByCurlReq{}
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
		rsp := schema.PublicRsp{}
		res := []byte{}
		fmt.Println(string(message))

		if currentParam.ConcurrencyQuantity == 0 || currentParam.PressureTime == 0 {
			// 必须先设置并发数和压测时间
			if err = json.Unmarshal(message, &currentParam); err != nil {
				fmt.Println("Error unmarshal message:" + err.Error())
				rsp.Code = 400
				rsp.Message = "请先设置并发数或压测时间，或设置失败检查参数"
			} else {
				rsp.Code = 200
				rsp.Message = "设置并发数/压测时间成功"
			}
			goto END
		}
		//识别修改并发数
		if err = json.Unmarshal(message, &currentParam); err == nil {
			rsp.Code = 200
			rsp.Message = "修改并发数/压测时间成功"
			goto END
		} else {
			if currentParam.ConcurrencyQuantity == 0 || currentParam.PressureTime == 0 {
				fmt.Println("Error ConcurrencyQuantity or PressureTime invalid")
				rsp.Code = 400
				rsp.Message = "Error ConcurrencyQuantity or PressureTime invalid"
			} else {
				// logic
				request, err := model.NewRequest("", verify, code, 0, false, string(message), headers, body, maxCon, http2, keepalive)
				if err != nil {
					fmt.Println(err.Error())
					rsp.Code = 400
					rsp.Message = err.Error()
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
						server.Dispose(c, currentParam.ConcurrencyQuantity, totalNumber, request, con, &w, startTime)
						currentParam.PressureTime--
					}
					w.Wait()
					rsp.Code = 0
				}
			}
		}
		fmt.Println(currentParam)

	END:
		res, _ = json.Marshal(rsp)
		err = con.WriteMessage(mt, res)
		if err != nil {
			fmt.Println("Error write message:" + err.Error())
			break
		}
	}

}
