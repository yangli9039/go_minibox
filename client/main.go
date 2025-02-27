package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	back_end "minibox/client/back-end"
	"time"
)

// 发布者10，升级中20，升级失败30,升级成功40
type TaskStatus int

const (
	DeviceStatus发布中  TaskStatus = 10
	DeviceStatus升级中  TaskStatus = 20
	DeviceStatus升级失败 TaskStatus = 20
	DeviceStatus升级成功 TaskStatus = 20
)

type Request struct {
	Action string      `json:"action"`
	Data   interface{} `json:"data"`
}

type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// 请求新任务
type JSONDataReq struct {
	Op         string      `json:"op"`
	DeviceInfo interface{} `json:"device_info"`
}

// 更新任务
type JSONDataUpdate struct {
	Op         string      `json:"op"`
	TaskID     int         `json:"task_id"`
	TaskStatus int         `json:"task_status"`
	DeviceInfo interface{} `json:"device_info"`
}

type PackageData struct {
	PackageName    string `json:"packageName"`
	PackageVersion string `json:"packageVersion"`
	PackageType    string `json:"packageType"`
	UpdatePolicy   string `json:"updatePolicy"`
	PackageURL     string `json:"packageUrl"`
	IsZip          bool   `json:"isZip"`
	PackageHash    string `json:"packageHash"`
	PackageSize    int    `json:"packageSize"`
}

type ExecData struct {
	Index       int    `json:"index"`
	Event       string `json:"event"`
	PackageName string `json:"packageName"`
	Args        struct {
		Path  string `json:"path"`
		Dist  string `json:"dist"`
		Count int    `json:"count"`
	} `json:"args"`
}

type JSONDataRsp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		DeviceID    string        `json:"device_id"`
		IsUpgrade   bool          `json:"is_upgrade"`
		TaskID      int           `json:"task_id"`
		PackageList []PackageData `json:"package_list"`
		ExecList    []ExecData    `json:"exec_list"`
	} `json:"data"`
}

func sendRequest(conn *websocket.Conn, action string, data interface{}) (JSONDataRsp, error) {

	request := Request{
		Data: data,
	}
	message, err := json.Marshal(request.Data)
	if err != nil {
		log.Fatal("Error marshalling message:", err)
	}

	if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
		log.Fatal("Error sending message:", err)
	}

	// 接收并打印服务端的响应
	_, response, err := conn.ReadMessage()
	if err != nil {
		log.Fatal("Error reading response:", err)
	}
	fmt.Printf("Received from server: %s\n", response)
	// true有新任务：有新任务执行后续处理。 false没有新任务：略过
	var serverResponse JSONDataRsp
	if err := json.Unmarshal(response, &serverResponse); err != nil {
		//log.Fatal("Error unmarshalling message:", err)
	}

	if serverResponse.Message == "success" {
		return serverResponse, nil
	} else {
		fmt.Printf("Received from server: %s\n", serverResponse.Message)
	}

	return serverResponse, errors.New("查询出错")
}

func main() {
	// 注册设备
	var Host string = "http://192.168.114.251:25088/v1"
	var path string = "/base_device/register"
	Url := Host + path //http://192.168.114.251:25088/v1/base_device/register
	register := back_end.NewRegister(Url, map[string]string{"DeviceSn": "iot001"})
	ws_url, err := back_end.ToRegister(register)
	if err != nil {
		log.Fatal("连接失败")
	}
	// 连接到 WebSocket 服务端
	conn, _, err := websocket.DefaultDialer.Dial(ws_url, nil)
	if err != nil {
		log.Fatal("Dial failed:", err)
	}
	defer conn.Close()

	// 使用 ticker 每 30 秒发送一次心跳
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// 定时发送心跳消息
	for {
		select {
		case <-ticker.C:
			/*
				{
				"op": "task_check",
				"device_info": "{\"cpu\":\"60%\"}"
				}
			*/
			info := map[string]interface{}{
				"cpu": "61%",
			}
			heartbeat := JSONDataReq{
				Op:         "task_check",
				DeviceInfo: info,
			}
			rsp, err := sendRequest(conn, "check", heartbeat)
			if err != nil {
				log.Println(err.Error())
			}
			if rsp.Data.IsUpgrade {
				// 回复管理平台更新中
				dic := map[string]interface{}{
					"op":          "task_status",
					"task_id":     rsp.Data.TaskID,
					"task_status": DeviceStatus升级中,
					"device_info": "",
				}
				rsp, err := sendRequest(conn, "update", dic)
				if err != nil {
					log.Printf("err")
				}

				// 解析数据
				//rsp.Data.PackageList.
			}

		}
	}

	//// 持续接收消息（包括心跳）
	//for {
	//	_, response, err := conn.ReadMessage()
	//	if err != nil {
	//		log.Fatal("Error reading message:", err)
	//	}
	//
	//	var serverResponse JSONDataRsp
	//	if err := json.Unmarshal(response, &serverResponse); err != nil {
	//		log.Fatal("Error unmarshalling message:", err)
	//	}
	//
	//	if serverResponse.Message == "success" {
	//		fmt.Println("Received heartbeat from server")
	//
	//	} else {
	//		fmt.Printf("Received from server: %s\n", serverResponse.Message)
	//	}
	//}
}
