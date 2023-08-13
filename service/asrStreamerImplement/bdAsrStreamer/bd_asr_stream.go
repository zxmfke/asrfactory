package bdAsrStreamer

import (
	"asrer/define"
	"asrer/log"
	"asrer/util"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"strings"
	"time"
)

// https://ai.baidu.com/ai-doc/SPEECH/2k5dllqxj

type BdAsrStreamer struct {
	Cfg    BdAsrConfig
	TaskID string
	conn   *websocket.Conn
	close  bool
	begin  bool
}

func NewBdAsrStreamer(taskID string, config BdAsrConfig) *BdAsrStreamer {
	return &BdAsrStreamer{
		TaskID: taskID,
		Cfg:    config,
	}
}

func (b *BdAsrStreamer) Init() error {
	var (
		err      error
		response *http.Response
	)

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	if b.conn, response, err = dialer.Dial("wss://vop.baidu.com/realtime_asr?sn="+b.TaskID, nil); err != nil {
		log.Errorf("dail ws err : %s", err.Error())
		return err
	}

	if response.StatusCode != 101 {
		log.Errorf("StatusCode not 101 ws err : %s", response.Status)
		return fmt.Errorf("%s", response.Status)
	}

	return b.start()
}

func (b *BdAsrStreamer) start() error {

	req := bdAsrReq{
		Type: Start,
		Data: bdReqData{
			AppID:  b.Cfg.AppID,
			AppKey: b.Cfg.AppKey,
			DevPID: 15372,
			Cuid:   util.GetUuid(),
			Format: "pcm",
			Sample: b.Cfg.SampleRate,
		},
	}

	data, _ := json.Marshal(req)

	b.begin = true

	return b.conn.WriteMessage(websocket.TextMessage, data)
}

func (b *BdAsrStreamer) Send(data []byte) error {
	if b.close {
		return fmt.Errorf("%s", "ws closed!")
	}

	if err := b.conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
		return err
	}

	return nil
}

func (b *BdAsrStreamer) Recv(output chan<- define.Output) {

	output <- define.Output{Status: define.Start}

	for {
		_, msg, err := b.conn.ReadMessage()
		if err != nil {

			if b.close {
				return
			}

			if strings.Contains(err.Error(), "close 1005") {
				output <- define.Output{Status: define.Done}
				return
			}

			log.Errorf("%s", err.Error())

			output <- define.Output{
				Err: err,
			}
			return
		}

		var (
			resp   bdAsrResult
			status       = define.SentencePartial
			eOut   error = nil
		)

		_ = json.Unmarshal(msg, &resp)

		if resp.ErrNo != 0 {

			if resp.ErrNo == -3005 {

				log.Errorf("startTime:%d , endTime:%d, err_no:-3005,%s", resp.StartTime, resp.EndTime, resp.ErrMsg)
				continue
			}

			if b.begin {
				status = define.Start
				b.begin = false
			}

			output <- define.Output{
				Status: status,
				Err:    fmt.Errorf("%d | %s", resp.ErrNo, resp.ErrMsg),
			}
			return
		}

		if resp.Type == HeartBeat {
			continue
		}

		if resp.Type == FinalText {
			status = define.SentenceFinal
			log.Infow("end payload", "begin time", resp.StartTime, "end time", resp.EndTime)
		}

		output <- define.Output{
			Status: status,
			Text:   resp.Result,
			Err:    eOut,
		}

		if status == define.Done {
			return
		}
	}
}

func (b *BdAsrStreamer) End() error {

	req := bdAsrReq{
		Type: Finish,
	}

	data, _ := json.Marshal(req)
	err := b.conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		log.Errorf("发送消息失败：", err)
		return err
	}

	return nil
}

func (b *BdAsrStreamer) Close() {
	_ = b.conn.Close()
	b.close = true
}
