package golink

import (
	"context"
	"strings"
	"sync"
	"time"

	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"pressure-test-admin/tool/go-stress-testing/helper"

	"pressure-test-admin/tool/go-stress-testing/model"
)

// Grpc grpc 接口请求
func Radius(ctx context.Context, chanID uint64, ch chan<- *model.RequestResults, totalNumber uint64, wg *sync.WaitGroup,
	request *model.Request) {
	defer func() {
		wg.Done()
	}()
	for i := uint64(0); i < totalNumber; i++ {
		authRequest(chanID, ch, i, request)
	}
	return
}

// grpcRequest 请求
func authRequest(chanID uint64, ch chan<- *model.RequestResults, i uint64, request *model.Request) {
	var (
		startTime = time.Now()
		isSucceed = false
		errCode   = int(radius.CodeAccessAccept)
	)
	// 需要发送的数据
	// fmt.Printf("rsp:%+v", rsp)
	packet := radius.New(radius.CodeAccessRequest, []byte(`cisco`))
	index := strings.Index(request.URL, "@")
	username := "tim"
	host := request.URL
	if index != -1 {
		username = username + "@" + request.URL[index+1:]
		host = request.URL[:index]
	}
	rfc2865.UserName_SetString(packet, username)
	rfc2865.UserPassword_SetString(packet, "12345678")
	rfc2865.NASPortType_Set(packet, rfc2865.NASPortType_Value_Ethernet)
	rfc2865.ServiceType_Set(packet, rfc2865.ServiceType_Value_FramedUser)
	rfc2865.NASIdentifier_Set(packet, []byte(`benchmark`))
	rsp, err := radius.Exchange(context.Background(), packet, host)
	if err != nil {
		errCode = model.RequestErr
	} else {
		if rsp.Code != radius.CodeAccessAccept {
			errCode = int(rsp.Code)
		} else {
			isSucceed = true
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
