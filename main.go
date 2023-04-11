package main

import (
	"math/rand"
	"pressure-test-admin/logic"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	// 1.创建路由
	r := gin.Default()
	// 2.绑定路由规则，执行的函数
	// gin.Context，封装了request和response
	r.GET("/pressureByUrl", logic.PressureByUrl)
	r.GET("/pressureByCurl", logic.PressureByCurl)
	r.GET("/pressureByGolang", logic.PressureByGolang)
	// 3.监听端口，默认在8080
	// Run("里面不指定端口号默认为8080")
	r.Run(":8000")
}
