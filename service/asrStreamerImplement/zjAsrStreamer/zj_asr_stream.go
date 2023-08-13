package zjAsrStreamer

import (
	"asrer/define"
	"asrer/log"
	"asrer/util"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"strings"
)

// https://www.volcengine.com/docs/6561/80818

type ZjAsrStreamer struct {
	Cfg    ZjAsrConfig
	TaskID string
	conn   *websocket.Conn
	close  bool
	begin  bool
	end    bool
}

func NewZjAsrStreamer(taskID string, config ZjAsrConfig) *ZjAsrStreamer {

	return &ZjAsrStreamer{
		TaskID: taskID,
		Cfg:    config,
	}
}

func (z *ZjAsrStreamer) Init() error {

	var (
		err      error
		response *http.Response
	)

	var tokenHeader = http.Header{"Authorization": []string{fmt.Sprintf("Bearer;%s", z.Cfg.Token)}}

	if z.conn, response, err = websocket.DefaultDialer.Dial("wss://openspeech.bytedance.com/api/v2/asr", tokenHeader); err != nil {
		return err
	}

	if response.StatusCode != 101 {
		log.Errorf("StatusCode not 101 ws err : %s", response.Status)
		return fmt.Errorf("%s", response.Status)
	}

	return z.start()
}

func (z *ZjAsrStreamer) start() error {
	req := z.constructRequest()
	payload := gzipCompress(req)
	payloadSize := len(payload)
	payloadSizeArr := make([]byte, 4)
	binary.BigEndian.PutUint32(payloadSizeArr, uint32(payloadSize))

	fullClientMsg := make([]byte, len(DefaultFullClientWsHeader))
	copy(fullClientMsg, DefaultFullClientWsHeader)
	fullClientMsg = append(fullClientMsg, payloadSizeArr...)
	fullClientMsg = append(fullClientMsg, payload...)

	z.begin = true

	return z.conn.WriteMessage(websocket.BinaryMessage, fullClientMsg)
}

func (z *ZjAsrStreamer) Send(data []byte) error {

	var (
		payload []byte
	)

	z.begin = false

	audioMsg := make([]byte, len(DefaultAudioOnlyWsHeader))
	copy(audioMsg, DefaultAudioOnlyWsHeader)

	payload = gzipCompress(data)
	payloadSize := len(payload)
	payloadSizeArr := make([]byte, 4)
	binary.BigEndian.PutUint32(payloadSizeArr, uint32(payloadSize))
	audioMsg = append(audioMsg, payloadSizeArr...)
	audioMsg = append(audioMsg, payload...)

	return z.conn.WriteMessage(websocket.BinaryMessage, audioMsg)
}

func (z *ZjAsrStreamer) Recv(output chan<- define.Output) {

	output <- define.Output{Status: define.Start}

	for {

		_, msg, err := z.conn.ReadMessage()

		if err != nil {

			if z.close {
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
			status = define.SentencePartial
		)

		asrResponse, err := z.parseResponse(msg)
		if err != nil {
			output <- define.Output{Status: define.Done, Err: err}
			return
		}

		//log.Infof("%d,%s", asrResponse.Code, asrResponse.Message
		//log.Infof("%+v", asrResponse)

		if asrResponse.Code != SuccessCode {
			if z.begin {
				status = define.Start
			}

			output <- define.Output{Status: status, Err: fmt.Errorf("%s", asrResponse.Message)}
			return
		}

		if z.end {
			output <- define.Output{Status: define.Done}
			return
		}

		if len(asrResponse.Results) == 0 || len(asrResponse.Results) > 1 {
			continue
		}

		tmp := asrResponse.Results[0]

		if len(tmp.Utterances) == 0 || len(tmp.Utterances) > 1 {
			continue
		}

		resp := tmp.Utterances[0]

		if resp.Text == "" {
			continue
		}

		if resp.Definite {
			status = define.SentenceFinal

			log.Infow("end payload", "start time", resp.StartTime, "end time", resp.EndTime)

		}

		output <- define.Output{
			Status: status,
			Text:   resp.Text,
		}
	}
}

func (z *ZjAsrStreamer) End() error {
	if z.close {
		return fmt.Errorf("%s", "ws closed!")
	}

	audioMsg := make([]byte, len(DefaultLastAudioWsHeader))
	copy(audioMsg, DefaultLastAudioWsHeader)
	payload := gzipCompress(nil)
	payloadSize := len(payload)
	payloadSizeArr := make([]byte, 4)
	binary.BigEndian.PutUint32(payloadSizeArr, uint32(payloadSize))
	audioMsg = append(audioMsg, payloadSizeArr...)
	audioMsg = append(audioMsg, payload...)

	z.end = true

	return z.conn.WriteMessage(websocket.BinaryMessage, audioMsg)
}

func (z *ZjAsrStreamer) Close() {
	if z.close {
		return
	}

	_ = z.conn.Close()
	return
}

func (z *ZjAsrStreamer) constructRequest() []byte {
	reqID := util.NewV4().String()
	req := make(map[string]map[string]interface{})
	req["app"] = make(map[string]interface{})
	req["app"]["appid"] = z.Cfg.AppID
	req["app"]["cluster"] = z.Cfg.Cluster
	req["app"]["token"] = z.Cfg.Token
	req["user"] = make(map[string]interface{})
	req["user"]["uid"] = "uid"
	req["request"] = make(map[string]interface{})
	req["request"]["reqid"] = reqID
	req["request"]["nbest"] = 1
	req["request"]["workflow"] = "audio_in,resample,partition,vad,fe,decode"
	req["request"]["result_type"] = "single"
	req["request"]["show_utterances"] = true
	req["request"]["sequence"] = 1
	req["audio"] = make(map[string]interface{})
	req["audio"]["format"] = z.Cfg.Format
	req["audio"]["codec"] = "raw"
	req["audio"]["rate"] = z.Cfg.SampleRate
	req["audio"]["language"] = z.Cfg.Language
	reqStr, _ := json.Marshal(req)
	return reqStr
}

func (z *ZjAsrStreamer) parseResponse(msg []byte) (AsrResponse, error) {
	//protocol_version := msg[0] >> 4
	headerSize := msg[0] & 0x0f
	messageType := msg[1] >> 4
	//message_type_specific_flags := msg[1] & 0x0f
	serializationMethod := msg[2] >> 4
	messageCompression := msg[2] & 0x0f
	//reserved := msg[3]
	//header_extensions := msg[4:header_size * 4]
	payload := msg[headerSize*4:]
	payloadMsg := make([]byte, 0)
	payloadSize := 0
	//print('message type: {}'.format(message_type))

	if messageType == byte(SERVER_FULL_RESPONSE) {
		payloadSize = int(int32(binary.BigEndian.Uint32(payload[0:4])))
		payloadMsg = payload[4:]
	} else if messageType == byte(SERVER_ACK) {
		seq := int32(binary.BigEndian.Uint32(payload[:4]))
		if len(payload) >= 8 {
			payloadSize = int(binary.BigEndian.Uint32(payload[4:8]))
			payloadMsg = payload[8:]
		}
		log.Infof("SERVER_ACK seq: %s", seq)
	} else if messageType == byte(SERVER_ERROR_RESPONSE) {
		code := int32(binary.BigEndian.Uint32(payload[:4]))
		payloadSize = int(binary.BigEndian.Uint32(payload[4:8]))
		payloadMsg = payload[8:]
		log.Errorf("SERVER_ERROR_RESPONSE code: %d", code)
		return AsrResponse{}, errors.New(string(payloadMsg))
	}
	if payloadSize == 0 {
		return AsrResponse{}, errors.New("payload size if 0")
	}
	if messageCompression == byte(GZIP) {
		payloadMsg = gzipDecompress(payloadMsg)
	}

	var asrResponse = AsrResponse{}
	if serializationMethod == byte(JSON) {
		err := json.Unmarshal(payloadMsg, &asrResponse)
		if err != nil {
			return AsrResponse{}, err
		}
	}

	return asrResponse, nil
}
