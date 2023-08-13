package aliAsrStreamer

import (
	"asrer/define"
	"asrer/log"
	"asrer/util"
	"encoding/json"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

// https://help.aliyun.com/document_detail/324262.htm?spm=a2c4g.432038.0.0.327574cbrQ3qQx#topic-2121083

type AliAsrStreamer struct {
	Cfg       AliAsrConfig
	TaskID    string
	Namespace string
	conn      *websocket.Conn
	close     bool
	begin     bool
}

func NewAliAsrStreamer(taskID string, config AliAsrConfig) *AliAsrStreamer {
	return &AliAsrStreamer{
		Namespace: "SpeechTranscriber",
		TaskID:    taskID,
		Cfg:       config,
	}
}

func (a *AliAsrStreamer) Init() error {
	var (
		err      error
		token    *AliTokenRes
		response *http.Response
	)

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	if token, err = a.GetToken(); err != nil {
		log.Errorf("get token err : %s", err.Error())
		return err
	}

	header := http.Header{
		"X-NLS-Token": []string{token.Token.Id},
	}

	if a.conn, response, err = dialer.Dial("wss://nls-gateway-cn-beijing.aliyuncs.com/ws/v1", header); err != nil {
		log.Errorf("dail ws err : %s", err.Error())
		return err
	}

	if response.StatusCode != DailOK {
		log.Errorf("StatusCode not 101 ws err : %s", response.Status)
		return fmt.Errorf("%s", response.Status)
	}

	return a.start()
}

func (a *AliAsrStreamer) start() error {

	req := a.CommonRequest(AsrStreamStart)
	payload, _ := json.Marshal(a.defaultStartTranscriptionParam("wav"))
	_ = json.Unmarshal(payload, &req.Payload)
	data, _ := json.Marshal(req)

	a.begin = true

	return a.conn.WriteMessage(websocket.TextMessage, data)
}

func (a *AliAsrStreamer) Send(data []byte) error {

	if a.close {
		return fmt.Errorf("%s", "ws closed!")
	}

	a.begin = false

	if err := a.conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
		return err
	}

	return nil
}

func (a *AliAsrStreamer) Recv(output chan<- define.Output) {

	//output <- define.Output{Status: define.Start}

	for {
		_, msg, err := a.conn.ReadMessage()
		if err != nil {

			if a.close {
				return
			}

			log.Errorf("ws read wrong:%s", err.Error())

			output <- define.Output{
				Err: err,
			}
			return
		}

		var (
			resp         = new(CommonRequest)
			status       = define.SentencePartial
			eOut   error = nil
		)

		_ = json.Unmarshal(msg, &resp)

		if resp.Header.Status != StatusOK {

			if a.begin {
				status = define.Start
			}

			output <- define.Output{
				Status: status,
				Err:    fmt.Errorf("%s", resp.Header.StatusText),
			}
			return
		}

		ot := define.Output{
			Err: eOut,
		}

		switch resp.Header.Name {
		case start:
			ot.Status = define.Start
			break
		case completed:
			ot.Status = define.Done
			break
		case sentenceEnd:
			ot.Status = define.SentenceFinal
			payload := new(TranscriptionResultEndResp)
			resp.Payload = payload
			_ = json.Unmarshal(msg, &resp)
			ot.Text = payload.Result
			//log.Infof("END %+v", *payload)
			break
		case sentenceOnChanged:
			ot.Status = define.SentencePartial
			payload := new(TranscriptionResultChangedResp)
			resp.Payload = payload
			_ = json.Unmarshal(msg, &resp)
			//log.Infof("PARTICIAL %+v", *payload)
			ot.Text = payload.Result
			break
		default:
			log.Infof("jump header name : %s", resp.Header.Name)
			break
		}

		if resp.Header.Name == sentenceBegin {
			continue
		}

		output <- ot

		if ot.Status == define.Done {
			return
		}
	}
}

func (a *AliAsrStreamer) End() error {

	data, _ := json.Marshal(a.CommonRequest(AsrStreamStop))

	if err := a.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return err
	}

	return nil
}

func (a *AliAsrStreamer) Close() {
	_ = a.conn.Close()
	a.close = true
}

func (a *AliAsrStreamer) CommonRequest(action string) *CommonRequest {
	req := new(CommonRequest)
	//req.Context = defaultContext
	req.Header.AppKey = a.Cfg.AppKey
	req.Header.Namespace = a.Namespace
	req.Header.TaskID = a.TaskID
	req.Header.MessageID = util.GetUuid()
	req.Header.Name = action
	req.Payload = nil

	return req
}

func (a *AliAsrStreamer) GetToken() (*AliTokenRes, error) {

	var (
		err      error
		revMsg   = new(AliTokenRes)
		client   *sdk.Client
		response *responses.CommonResponse
	)

	if client, err = sdk.NewClientWithAccessKey(a.Cfg.Region, a.Cfg.AppID, a.Cfg.AppSecret); err != nil {
		return nil, err
	}

	request := requests.NewCommonRequest()
	request.Method = "POST"
	request.Domain = a.Cfg.Domain
	request.ApiName = "CreateToken"
	request.Version = a.Cfg.Version

	if response, err = client.ProcessCommonRequest(request); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(response.GetHttpContentBytes(), revMsg); err != nil {
		return nil, err
	}

	return revMsg, err
}
