package back_end

import (
	"errors"
	"fmt"
	"log"
	"resty.dev/v3"
	"strings"
)

/*
{
"DeviceSn": "cx10"
}
=>
{
	"code": 0,
	"data": {
		"device_id": "343672220331016192",
		"ws_url": "ws://192.168.114.251:25088/v1/base_device/heartbeat/ws?device_id=343672220331016192"
	},
	"msg": "成功"
}
<>
{
	"code": 7,
	"data": {},
	"msg": "创建失败:Error 1062 (23000): Duplicate entry 'cx10' for key 'base_device.idx_base_device_device_sn'"
}

*/

type Register struct {
	Url    string            `json:"url"`
	Params map[string]string `json:"params"`
}

type RegisterResponse struct {
	Code int `json:"code"`
	Data struct {
		DeviceID string `json:"device_id"`
		WsURL    string `json:"ws_url"`
	} `json:"data"`
	Msg string `json:"msg"`
}

type RegisterError struct {
	Code int      `json:"code"`
	Data struct{} `json:"data"`
	Msg  string   `json:"msg"`
}

func NewRegister(url string, params map[string]string) *Register {
	return &Register{url, params}
}

func ToRegister(register *Register) (ws_url string, err error) {
	client := resty.New()
	defer client.Close()

	var registerResp RegisterResponse
	var registerErr RegisterError
	res, err := client.R().
		//SetBody(map[string]string{
		//	"DeviceSn": "iot001",
		//}).
		SetBody(register.Params).
		SetResult(&registerResp).
		SetError(&registerErr).
		Post(register.Url)

	fmt.Println(err, res)
	if len(registerResp.Data.WsURL) > 0 {
		return registerResp.Data.WsURL, nil
	}

	if strings.Contains(registerErr.Msg, "Duplicate") {
		return "", errors.New("重复注册")
	}
	if err != nil {
		log.Fatal(err)
	}
	log.Println(res)
	return "", nil
}
