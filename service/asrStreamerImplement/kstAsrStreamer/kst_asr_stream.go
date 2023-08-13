package kstAsrStreamer

import (
	"asrer/define"
	"asrer/log"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"strings"
	"time"
)

// https://aihc.shengwenyun.com/asr-stream-md

type KstAsrStreamer struct {
	Cfg    KstAsrConfig
	TaskID string
	conn   *websocket.Conn
	close  bool
	begin  bool
}

func NewKstAsrStreamer(taskID string, config KstAsrConfig) *KstAsrStreamer {
	return &KstAsrStreamer{
		TaskID: taskID,
		Cfg:    config,
		begin:  true,
	}
}

func (k *KstAsrStreamer) Init() error {

	var (
		response *http.Response
	)

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	token, err := k.GetToken()
	if err != nil {
		return err
	}

	header := http.Header{
		"Sec-WebSocket-Protocol": []string{token},
	}

	if k.conn, response, err = dialer.Dial(k.Cfg.url(), header); err != nil {
		log.Errorf("dail ws err : %s", err.Error())
		return err
	}

	if response.StatusCode != 101 {
		log.Errorf("StatusCode not 101 ws err : %s", response.Status)
		return fmt.Errorf("%s", response.Status)
	}

	return nil
}

func (k *KstAsrStreamer) Send(data []byte) error {

	if k.close {
		return fmt.Errorf("%s", "ws closed!")
	}

	if err := k.conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
		return err
	}

	return nil
}

func (k *KstAsrStreamer) Recv(output chan<- define.Output) {

	if k.begin {
		output <- define.Output{Status: define.Start}
		k.begin = false
	}

	for {
		_, msg, err := k.conn.ReadMessage()

		if err != nil {

			if k.close {
				return
			}

			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				output <- define.Output{Err: fmt.Errorf("websocket 连接已关闭")}
				break
			}

			output <- define.Output{Err: fmt.Errorf("读取失败:%s", err.Error())}
			break
		}

		var (
			eOut   error = nil
			status       = define.SentencePartial
		)

		resp := new(WsResp)
		if err := json.Unmarshal(msg, resp); err != nil {
			output <- define.Output{
				Err: fmt.Errorf("读取ws返回信息反序列化错误:%s", err.Error()),
			}
			return
		}

		if resp.Code != 0 {
			output <- define.Output{
				Err: fmt.Errorf("读取ws返回信息Code不为0:%s", resp.Msg),
			}
			return
		}

		switch resp.Data.Typ {
		case 1:
			status = define.SentencePartial
		case 2:
			status = define.SentenceFinal
			log.Infow("end payload", "begin time", resp.Data.Ss, "end time", resp.Data.Se)
		}

		if resp.Data.Result == "end" {
			status = define.Done
		}

		output <- define.Output{
			Status: status,
			Text:   resp.Data.Result,
			Err:    eOut,
		}

		if status == define.Done {
			return
		}
	}
}

func (k *KstAsrStreamer) End() error {

	closeMode := WSMode{Mode: "CLOSE"}
	closeJson, _ := json.Marshal(closeMode)

	err := k.conn.WriteMessage(websocket.TextMessage, closeJson)
	if err != nil {
		log.Errorf("发送消息失败：", err)
		return err
	}

	return nil
}

func (k *KstAsrStreamer) Close() {
	_ = k.conn.Close()
	k.close = true
}

func (k *KstAsrStreamer) GetToken() (string, error) {

	var (
		url     = "https://aihc.shengwenyun.com/aihc/auth"
		method  = "POST"
		payload = strings.NewReader(`{"app_id": "` + k.Cfg.AppID + `","app_secret": "` + k.Cfg.AppSecret + `"}`)
		err     error
		req     *http.Request
		res     *http.Response
	)

	client := &http.Client{}

	if req, err = http.NewRequest(method, url, payload); err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", "application/json")

	if res, err = client.Do(req); err != nil {
		return "", err
	}

	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("not 200 but : %d", res.StatusCode)
	}

	data, _ := io.ReadAll(res.Body)

	tokenResp := new(Resp)

	_ = json.Unmarshal(data, tokenResp)

	if tokenResp.Code != 0 {
		return "", fmt.Errorf("%s", tokenResp.Msg)
	}

	return tokenResp.Data.AccessToken, nil
}
