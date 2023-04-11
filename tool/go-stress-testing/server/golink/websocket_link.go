// Package golink 连接
package golink

import (
	"context"
	"fmt"
	"sync"
	"time"

	"pressure-test-admin/tool/go-stress-testing/helper"
	"pressure-test-admin/tool/go-stress-testing/model"
	"pressure-test-admin/tool/go-stress-testing/server/client"
)

const (
	firstTime    = 1 * time.Second // 连接以后首次请求数据的时间
	intervalTime = 1 * time.Second // 发送数据的时间间隔
)

var (
	// 请求完成以后是否保持连接
	keepAlive bool
)

func init() {
	keepAlive = true
}

// WebSocket webSocket go link
func WebSocket(ctx context.Context, chanID uint64, ch chan<- *model.RequestResults, totalNumber uint64,
	wg *sync.WaitGroup, request *model.Request, ws *client.WebSocket) {
	defer func() {
		wg.Done()
	}()
	defer func() {
		_ = ws.Close()
	}()

	var (
		i uint64
	)
	// 本项目以压测最大连接数为目的，启用定时器是为了连接保活
	// 不需要一直发送消息，如果需要压测 QPS 为目的，则需要取消定时器功能
	t := time.NewTimer(firstTime)
	for {
		select {
		case <-t.C:
			t.Reset(intervalTime)
			// 请求
			webSocketRequest(chanID, ch, i, request, ws)
			// 结束条件
			i = i + 1
			if i >= totalNumber {
				goto end
			}
		}
	}
end:
	t.Stop()

	if keepAlive {
		// 保持连接
		chWaitFor := make(chan int, 0)
		<-chWaitFor
	}
	return
}

// webSocketRequest 请求
func webSocketRequest(chanID uint64, ch chan<- *model.RequestResults, i uint64, request *model.Request,
	ws *client.WebSocket) {
	var (
		startTime = time.Now()
		isSucceed = false
		errCode   = model.HTTPOk
		msg       []byte
	)
	// 需要发送的数据
	seq := fmt.Sprintf("%d_%d", chanID, i)
	err := ws.Write([]byte(`{"seq":"` + seq + `","cmd":"ping","data":{}}`))
	if err != nil {
		errCode = model.RequestErr // 请求错误
	} else {
		msg, err = ws.Read()
		if err != nil {
			errCode = model.ParseError
			fmt.Println("读取数据 失败~")
		} else {
			errCode, isSucceed = request.GetVerifyWebSocket()(request, seq, msg)
		}
	}
	requestTime := uint64(helper.DiffNano(startTime))
	requestResults := &model.RequestResults{
		Time:      requestTime,
		IsSucceed: isSucceed,
		ErrCode:   errCode,
	}
	requestResults.SetID(chanID, i)
	ch <- requestResults
}
