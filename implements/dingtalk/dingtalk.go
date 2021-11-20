package dingtalk

import (
	"bytes"
	"context"
	"errors"
	jsoniter "github.com/json-iterator/go"
	"github.com/mutou1225/go-frame/config"
	"github.com/mutou1225/go-frame/logger"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

type DingTalkToken int

const (
	IgnoreToken  DingTalkToken = iota
	AllToken                   // 全部
	OperateToken               // 价格运营群
	OperateConf                // 价格运营配置后台
	Develop                    // 开发群
	OptAndDev                  // 价格运营群+开发群

	DingTalkUrl = "https://oapi.dingtalk.com/robot/send?access_token="
)

type DingTokenTextMsg struct {
	MsgType string `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	} `json:"text"`
	At struct {
		IsAtAll   bool     `json:"isAtAll"`
		AtMobiles []string `json:"atMobiles"`
	} `json:"at"`
}

// 提供的操作
type DingTalkOpt interface {
	// 发送信息到自定义的token
	DingTalkTextMsgWithToken(msg string, token string, isAtAll bool, mobiles ...string) error
	// 发送信息到配置的token
	DingTalkTextMsg(msg string, token DingTalkToken, isAtAll bool, mobiles ...string) error
}

// 发送钉钉Text信息
func DingTalkTextMsgWithToken(msg string, token string, isAtAll bool, mobiles ...string) error {
	/*if config.IsTest() {
		logger.PrintInfo("DingTalkTextMsg() Test Env IgnoreToken")
		return nil
	}*/

	if token == "" {
		return errors.New("token empty.")
	}

	textMsg := DingTokenTextMsg{}
	textMsg.MsgType = "text"
	textMsg.Text.Content = msg
	textMsg.At.IsAtAll = isAtAll
	textMsg.At.AtMobiles = mobiles

	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	jsonStr, err := json.Marshal(textMsg)
	if err != nil {
		return err
	}

	go func() {
		strUrl := DingTalkUrl + token
		logger.PrintInfo("curl -H'Content-Type:application/json' -d'%s' %s", string(jsonStr), strUrl)

		ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
		req, err := http.NewRequestWithContext(ctx, "POST", strUrl, bytes.NewReader(jsonStr))
		if err != nil {
			logger.PrintError("DingTalkTextMsg() Err: %s", err.Error())
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Charset", "utf-8")

		client := &http.Client{Timeout: 10 * time.Second}
		rsp, err := client.Do(req)
		if err != nil {
			logger.PrintError("DingTalkTextMsg() Err: %s", err.Error())
			return
		}
		defer rsp.Body.Close()

		if rsp.StatusCode != http.StatusOK {
			io.Copy(ioutil.Discard, rsp.Body)
			logger.PrintError("DingTalkTextMsg() Err: StatusCode: %d", rsp.StatusCode)
			return
		}

		body, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			logger.PrintError("DingTalkTextMsg() Err: %s", err.Error())
			return
		}

		logger.PrintInfo("DingTalkTextMsg() Ret: %s", string(body))
	}()

	return nil
}

// 发送钉钉Text信息
func DingTalkTextMsg(msg string, token DingTalkToken, isAtAll bool, mobiles ...string) error {
	if token == IgnoreToken {
		logger.PrintInfo("DingTalkTextMsg() IgnoreToken")
		return nil
	}

	tokenList := []string{}
	switch token {
	case AllToken:
		tokenList = append(tokenList, config.GetDingTalkOperate(), config.GetDingTalkOperateConf(), config.GetDingTalkDevelop())
	case OperateToken:
		tokenList = append(tokenList, config.GetDingTalkOperate())
	case OperateConf:
		tokenList = append(tokenList, config.GetDingTalkOperateConf())
	case Develop:
		tokenList = append(tokenList, config.GetDingTalkDevelop())
	case OptAndDev:
		tokenList = append(tokenList, config.GetDingTalkDevelop(), config.GetDingTalkOperate())
	}

	for _, token := range tokenList {
		err := DingTalkTextMsgWithToken(msg, token, isAtAll, mobiles...)
		if err != nil {
			logger.PrintInfo("DingTalkTextMsg() Err: %s", err.Error())
		}
	}

	return nil
}
