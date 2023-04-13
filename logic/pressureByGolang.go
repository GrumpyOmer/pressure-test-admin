package logic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"go/format"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"pressure-test-admin/schema"
	"pressure-test-admin/tool/go-stress-testing/model"
	"pressure-test-admin/tool/go-stress-testing/server"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func PressureByGolang(c *gin.Context) {
	con, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("Error during connection upgradation:" + err.Error())
		return
	}
	defer con.Close()
	currentParam := schema.PressureByGolangReq{}
	dir := 0
	for {
		// road request message
		mt, message, err := con.ReadMessage()
		if err != nil {
			fmt.Println("Error road message:" + err.Error())
			break
		}
		rsp := schema.PublicRsp{}
		res := []byte{}
		fmt.Println(string(message))

		if currentParam.ConcurrencyQuantity == 0 || currentParam.Port == 0 {
			// 必须先设置并发数和端口
			if err = json.Unmarshal(message, &currentParam); err != nil {
				fmt.Println("Error unmarshal message:" + err.Error())
				rsp.Code = 400
				rsp.Message = "请先设置并发数或进程使用端口，或设置失败检查参数"
			} else {
				rsp.Code = 200
				rsp.Message = "设置并发数/端口成功"
			}
			goto END
		}
		fmt.Println(currentParam)
		//识别修改并发数
		if err = json.Unmarshal(message, &currentParam); err == nil {
			rsp.Code = 200
			rsp.Message = "修改并发数/端口成功"
			goto END
		} else {
			if currentParam.ConcurrencyQuantity == 0 || currentParam.Port == 0 {
				fmt.Println("Error ConcurrencyQuantity or port invalid")
				rsp.Code = 400
				rsp.Message = "Error ConcurrencyQuantity or port invalid"
			} else {
				if ok := ScanPort("tcp", "127.0.0.1", currentParam.Port); ok {
					rsp.Code = 400
					rsp.Message = "当前端口已被使用，请更换"
					goto END
				}
				// logic
				if dir == 0 {
					// 生成子目录
					dir = rand.Intn(60000)
				}
				// 先解析并存储golang文件
				if err = ParseGolangCode(string(message), strconv.Itoa(dir)); err != nil {
					rsp.Code = 400
					rsp.Message = "Error ParseGolangCode invalid: " + err.Error()
					goto END
				}
				// 将golang代码文件起成一个服务以提供调用
				// 启动golang服务进程
				cmd := exec.Command("go", "run", "./tmp/"+strconv.Itoa(dir)+"/main.go")
				// 子进程输出打印在主线程控制台
				cmd.Stdout = os.Stdout
				err = cmd.Start()
				if err != nil {
					rsp.Code = 400
					rsp.Message = "Error RunGoService invalid: " + err.Error()
					goto END
				}
				request, err := model.NewRequest("127.0.0.1:"+strconv.Itoa(currentParam.Port), verify, code, 0, false, "", headers, body, maxCon, http2, keepalive)
				if err != nil {
					fmt.Println(err.Error())
					rsp.Code = 400
					rsp.Message = err.Error()
				} else {
					fmt.Printf("\n 开始启动  并发数:%d 请求数:%d 请求参数: \n", currentParam.ConcurrencyQuantity, totalNumber)
					request.Print()
					// 开始处理
					server.Dispose(c, currentParam.ConcurrencyQuantity, totalNumber, request, con)
					os.RemoveAll("./tmp/" + strconv.Itoa(dir) + "/")
					// 杀死当前启动的服务的进程
					exec.Command("taskkill", "/pid", strconv.Itoa(portInUse(currentParam.Port)), "-t", "-f").Run()
					continue
				}
			}
		}

	END:
		fmt.Println(currentParam)
		res, _ = json.Marshal(rsp)
		err = con.WriteMessage(mt, res)
		if err != nil {
			fmt.Println("Error write message:" + err.Error())
			break
		}
	}
}

// 解析并检测golang代码
func ParseGolangCode(code string, dir string) error {
	var newTmpDir = "./tmp/" + dir + "/"
	var codeBytes = []byte(code)
	// 格式化输出的代码
	if formatCode, err := format.Source(codeBytes); nil == err {
		// 格式化失败，就还是用 content 吧
		codeBytes = formatCode
	}
	fmt.Println(string(codeBytes))
	// 创建目录
	if err := os.Mkdir(newTmpDir, os.ModePerm); nil != err {
		os.RemoveAll(newTmpDir)
		fmt.Println(err)
		return err
	}
	// 创建文件
	tmpFile, err := os.Create(newTmpDir + "main.go")
	if err != nil {
		os.RemoveAll(newTmpDir)
		fmt.Println(err)
		return err
	}
	// 代码写入文件
	tmpFile.Write(codeBytes)
	tmpFile.Close()
	// 运行检测代码
	cmd := exec.Command("go", "build", tmpFile.Name())
	res, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
		fmt.Println(string(res))
		// 删除文件夹以及main文件
		os.RemoveAll(newTmpDir)
		return err
	}
	return nil
}

// 检测端口是否被使用
func ScanPort(protocol string, hostname string, port int) bool {
	fmt.Printf("scanning port %d \n", port)
	p := strconv.Itoa(port)
	addr := net.JoinHostPort(hostname, p)
	conn, err := net.DialTimeout(protocol, addr, 3*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

func portInUse(portNumber int) int {
	res := -1
	var outBytes bytes.Buffer
	cmdStr := fmt.Sprintf("netstat -ano -p tcp | findstr %d", portNumber)
	cmd := exec.Command("cmd", "/c", cmdStr)
	cmd.Stdout = &outBytes
	cmd.Run()
	resStr := outBytes.String()
	r := regexp.MustCompile(`\s\d+\s`).FindAllString(resStr, -1)
	if len(r) > 0 {
		pid, err := strconv.Atoi(strings.TrimSpace(r[0]))
		if err != nil {
			res = -1
		} else {
			res = pid
		}
	}
	return res
}
