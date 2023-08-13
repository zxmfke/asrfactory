package txAsrStreamer

import (
	"asrer/define"
	"asrer/log"
	"asrer/util"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// https://cloud.tencent.com/document/product/1093/48982

type TxAsrStreamer struct {
	Cfg    TxAsrConfig
	TaskID string
	conn   *websocket.Conn
	close  bool
	begin  bool
}

func NewTxAsrStreamer(taskID string, config TxAsrConfig) *TxAsrStreamer {
	return &TxAsrStreamer{
		TaskID: taskID,
		Cfg:    config,
		begin:  true,
	}
}

func (t *TxAsrStreamer) Init() error {

	var (
		err      error
		response *http.Response
	)

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	if t.conn, response, err = dialer.Dial(t.WsUrl(), nil); err != nil {
		log.Errorf("dail ws err : %s", err.Error())
		return err
	}

	if response.StatusCode != 101 {
		log.Errorf("StatusCode not 101 ws err : %s", response.Status)
		return fmt.Errorf("%s", response.Status)
	}

	return nil
}

func (t *TxAsrStreamer) WsUrl() string {

	model := "16k_zh"

	if t.Cfg.SampleRate == define.SampleRate8k {
		model = "8k_zh"
	}

	now := time.Now()
	ts := strconv.FormatInt(now.Unix(), 10)
	expired := strconv.FormatInt(now.AddDate(0, 0, 2).Unix(), 10)
	tmpURL := "asr.cloud.tencent.com/asr/v2/" + t.Cfg.AppID +
		"?engine_model_type=" + model +
		"&expired=" + expired +
		"&needvad=1" +
		"&nonce=" + ts +
		"&secretid=" + t.Cfg.AppSecretID +
		"&timestamp=" + ts +
		"&vad_silence_time=800" +
		"&voice_format=12" +
		"&voice_id=" + util.GetUuid()

	return "wss://" + tmpURL + "&signature=" + t.getSignature(tmpURL)
}

func (t *TxAsrStreamer) getSignature(origin string) string {
	mac := hmac.New(sha1.New, []byte(t.Cfg.AppKey))
	mac.Write([]byte(origin))
	return url.QueryEscape(base64.StdEncoding.EncodeToString(mac.Sum(nil)))
}

func (t *TxAsrStreamer) Send(data []byte) error {

	if t.close {
		return fmt.Errorf("%s", "ws closed!")
	}

	if err := t.conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
		return err
	}

	return nil
}

func (t *TxAsrStreamer) Recv(output chan<- define.Output) {

	if t.begin {
		t.begin = false
		output <- define.Output{Status: define.Start}
	}

	for {
		_, msg, err := t.conn.ReadMessage()
		if err != nil {

			if t.close {
				return
			}

			output <- define.Output{Err: fmt.Errorf("读取失败 : %s", err.Error())}
			return
		}

		var (
			resp   txAsrResp
			eOut   error = nil
			status       = define.SentencePartial
		)

		if err = json.Unmarshal(msg, &resp); err != nil {
			output <- define.Output{
				Err: fmt.Errorf("parse resp error: %s", err.Error()),
			}
			return
		}

		if resp.Code != 0 {
			output <- define.Output{Err: fmt.Errorf("%s", resp.Message)}
			return
		}

		// 最终结果
		if resp.Result.SliceType == 2 {
			status = define.SentenceFinal
			//log.Infow("end payload", "start time", resp.Result.StartTime, "end time", resp.Result.EndTime)
		}

		if resp.Final == 1 {
			status = define.Done
		}

		output <- define.Output{
			Status: status,
			Text:   resp.Result.VoiceTextStr,
			Err:    eOut,
		}

		if status == define.Done {
			return
		}
	}
}

func (t *TxAsrStreamer) End() error {

	if err := t.conn.WriteMessage(websocket.TextMessage, []byte("{\"type\": \"end\"}")); err != nil {
		return err
	}

	return nil
}

func (t *TxAsrStreamer) Close() {
	_ = t.conn.Close()
	t.close = true
}
