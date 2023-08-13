package xfAsrStreamer

import (
	"asrer/define"
	"asrer/log"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// https://www.xfyun.cn/doc/asr/rtasr/API.html

type XfAsrStreamer struct {
	Cfg    XfAsrConfig
	TaskID string
	conn   *websocket.Conn
	close  bool
}

func NewXfAsrStreamer(taskID string, config XfAsrConfig) *XfAsrStreamer {

	return &XfAsrStreamer{
		TaskID: taskID,
		Cfg:    config,
	}
}

func (x *XfAsrStreamer) Init() error {

	var (
		err      error
		response *http.Response
	)

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	if x.conn, response, err = dialer.Dial(x.WsUrl(), nil); err != nil {
		log.Errorf("dail ws err : %s", err.Error())
		return err
	}

	if response.StatusCode != 101 {
		log.Errorf("StatusCode not 101 ws err : %s", response.Status)
		return fmt.Errorf("%s", response.Status)
	}

	return nil
}

func (x *XfAsrStreamer) Send(data []byte) error {
	if x.close {
		return fmt.Errorf("%s", "ws closed!")
	}

	if err := x.conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
		return err
	}

	return nil
}

func (x *XfAsrStreamer) Recv(output chan<- define.Output) {

	for {

		_, msg, err := x.conn.ReadMessage()

		if err != nil {

			if x.close {
				return
			}

			if strings.Contains(err.Error(), "close 1000") {
				output <- define.Output{Status: define.Done}
				return
			}

			log.Errorf("读取失败", err.Error())
			output <- define.Output{Err: fmt.Errorf("读取失败 : %s", err.Error())}
			return
		}

		var (
			result map[string]string
			eOut   error = nil
			status       = define.SentencePartial
			resp   AsrResult
		)

		if err = json.Unmarshal(msg, &result); err != nil {
			output <- define.Output{
				Err: fmt.Errorf("parse result error: %s", err.Error()),
			}
			return
		}

		if result["code"] != "0" {
			output <- define.Output{Err: fmt.Errorf("invalid result: %s", string(msg))}
			return
		}

		if result["action"] == "started" {
			output <- define.Output{Status: define.Start}
			continue
		}

		if err = json.Unmarshal([]byte(result["data"]), &resp); err != nil {
			output <- define.Output{
				Err: fmt.Errorf("parse resp error: %s", err.Error()),
			}
			return
		}

		var (
			text = ""
		)

		for _, wse := range resp.Cn.St.Rt[0].Ws {
			for _, cwe := range wse.Cw {

				if cwe.W == "" {
					//print("静音")
					continue
				}

				text += cwe.W
			}
		}

		// 最终结果
		if resp.Cn.St.Type == "0" {
			status = define.SentenceFinal

			log.Infow("end payload", "start time", resp.Cn.St.Bg, "end time", resp.Cn.St.Ed)
		}

		output <- define.Output{
			Status: status,
			Text:   text,
			Err:    eOut,
		}
	}
}

func (x *XfAsrStreamer) End() error {

	err := x.conn.WriteMessage(websocket.BinaryMessage, []byte("{\"end\": true}"))
	if err != nil {
		log.Errorf("发送消息失败：", err)
		return err
	}

	return nil
}

func (x *XfAsrStreamer) Close() {
	_ = x.conn.Close()
	x.close = true
}

func (x *XfAsrStreamer) WsUrl() string {
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	return "wss://rtasr.xfyun.cn/v1/ws?appid=" + x.Cfg.AppID + "&ts=" + ts + "&signa=" + x.getSignature(ts)
}

func (x *XfAsrStreamer) getSignature(ts string) string {
	mac := hmac.New(sha1.New, []byte(x.Cfg.AppKey))
	strByte := []byte(x.Cfg.AppID + ts)
	strMd5Byte := md5.Sum(strByte)
	strMd5 := fmt.Sprintf("%x", strMd5Byte)
	mac.Write([]byte(strMd5))
	return url.QueryEscape(base64.StdEncoding.EncodeToString(mac.Sum(nil)))
}
