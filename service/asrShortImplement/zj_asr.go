package asrShortImplement

import (
	"asrer/define"
	"asrer/log"
	"asrer/util"
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type ZjAsrConfig struct {
	AppID   string
	Token   string
	Cluster string

	format     string
	sampleRate int
	language   string
}

type ZjAsr struct {
	Cfg    *ZjAsrConfig
	TaskID string
	conn   *websocket.Conn
}

func NewZjAsr(taskID string, cfg *ZjAsrConfig) *ZjAsr {
	return &ZjAsr{Cfg: cfg, TaskID: taskID}
}

func (z *ZjAsr) Do(fileName string, sampleRate define.AudioSampleRate) (string, error) {
	var (
		err         error
		audioData   []byte
		asrResponse AsrResponse
		result      = ""
	)

	z.Cfg.format = filepath.Ext(fileName)[1:]
	z.Cfg.sampleRate = sampleRate
	z.Cfg.language = "zh-CN"

	if audioData, err = os.ReadFile(fileName); err != nil {
		return "", fmt.Errorf("fail to read audio file: %s", err.Error())
	}

	if asrResponse, err = z.requestAsr(audioData); err != nil {
		return "", fmt.Errorf("fail to request asr, %s", err.Error())
	}

	for i := 0; i < len(asrResponse.Results); i++ {
		result += asrResponse.Results[i].Text
	}

	return result, nil
}

type ProtocolVersion byte
type MessageType byte
type MessageTypeSpecificFlags byte
type SerializationType byte
type CompressionType byte

const (
	// Message Type:
	serverFullResponse  = MessageType(0b1001)
	serverAck           = MessageType(0b1011)
	serverErrorResponse = MessageType(0b1111)

	JSON = SerializationType(0b0001)
	GZIP = CompressionType(0b0001)
)

var DefaultFullClientWsHeader = []byte{0x11, 0x10, 0x11, 0x00}
var DefaultLastAudioWsHeader = []byte{0x11, 0x22, 0x11, 0x00}

func gzipCompress(input []byte) []byte {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	_, _ = w.Write(input)
	_ = w.Close()
	return b.Bytes()
}

func gzipDecompress(input []byte) []byte {
	b := bytes.NewBuffer(input)
	r, _ := gzip.NewReader(b)
	out, _ := io.ReadAll(r)
	_ = r.Close()
	return out
}

type AsrResponse struct {
	ReqID    string     `json:"reqid"`
	Code     int        `json:"code"`
	Message  string     `json:"message"`
	Sequence int        `json:"sequence"`
	Results  []ZjResult `json:"result,omitempty"`
}

type ZjResult struct {
	// required
	Text       string `json:"text"`
	Confidence int    `json:"confidence"`
	// if show_language == true
	Language string `json:"language,omitempty"`
	// if show_utterances == true
	Utterances []Utterance `json:"utterances,omitempty"`
}

type Utterance struct {
	Text      string `json:"text"`
	StartTime int    `json:"start_time"`
	EndTime   int    `json:"end_time"`
	Definite  bool   `json:"definite"`
	Words     []Word `json:"words"`
	// if show_language = true
	Language string `json:"language"`
}

type Word struct {
	Text      string `json:"text"`
	StartTime int    `json:"start_time"`
	EndTime   int    `json:"end_time"`
	Pronounce string `json:"pronounce"`
	// in docs example - blank_time
	BlankDuration int `json:"blank_duration"`
}

type WsHeader struct {
	ProtocolVersion          ProtocolVersion
	DefaultHeaderSize        int
	MessageType              MessageType
	MessageTypeSpecificFlags MessageTypeSpecificFlags
	SerializationType        SerializationType
	CompressionType          CompressionType
}

func (z *ZjAsr) requestAsr(audioData []byte) (AsrResponse, error) {
	var (
		err         error
		tokenHeader = http.Header{"Authorization": []string{fmt.Sprintf("Bearer;%s", z.Cfg.Token)}}
	)

	if z.conn, _, err = websocket.DefaultDialer.Dial("wss://openspeech.bytedance.com/api/v2/asr", tokenHeader); err != nil {
		return AsrResponse{}, err
	}
	defer func() {
		_ = z.conn.Close()
	}()

	// 1. send full z request
	if err = z.start(); err != nil {
		return AsrResponse{}, nil
	}

	return z.do(audioData)
}

func (z *ZjAsr) start() error {
	req := z.constructRequest()
	payload := gzipCompress(req)
	payloadSize := len(payload)
	payloadSizeArr := make([]byte, 4)
	binary.BigEndian.PutUint32(payloadSizeArr, uint32(payloadSize))

	fullClientMsg := make([]byte, len(DefaultFullClientWsHeader))
	copy(fullClientMsg, DefaultFullClientWsHeader)
	fullClientMsg = append(fullClientMsg, payloadSizeArr...)
	fullClientMsg = append(fullClientMsg, payload...)
	_ = z.conn.WriteMessage(websocket.BinaryMessage, fullClientMsg)
	_, msg, err := z.conn.ReadMessage()
	if err != nil {
		return err
	}

	if _, err = z.parseResponse(msg); err != nil {
		return err
	}

	return nil
}

func (z *ZjAsr) do(audioData []byte) (AsrResponse, error) {
	audioMsg := make([]byte, len(DefaultLastAudioWsHeader))
	copy(audioMsg, DefaultLastAudioWsHeader)
	payload := gzipCompress(audioData)
	payloadSize := len(payload)
	payloadSizeArr := make([]byte, 4)
	binary.BigEndian.PutUint32(payloadSizeArr, uint32(payloadSize))
	audioMsg = append(audioMsg, payloadSizeArr...)
	audioMsg = append(audioMsg, payload...)
	_ = z.conn.WriteMessage(websocket.BinaryMessage, audioMsg)
	_, msg, err := z.conn.ReadMessage()
	if err != nil {
		return AsrResponse{}, err
	}

	asrResponse, err := z.parseResponse(msg)
	if err != nil {
		return AsrResponse{}, err
	}

	return asrResponse, nil
}

func (z *ZjAsr) constructRequest() []byte {
	req := make(map[string]map[string]interface{})
	req["app"] = make(map[string]interface{})
	req["app"]["appid"] = z.Cfg.AppID
	req["app"]["cluster"] = z.Cfg.Cluster
	req["app"]["token"] = z.Cfg.Token
	req["user"] = make(map[string]interface{})
	req["user"]["uid"] = "uid"
	req["request"] = make(map[string]interface{})
	req["request"]["reqid"] = util.GetUuid()
	req["request"]["nbest"] = 1
	req["request"]["workflow"] = "audio_in,resample,partition,vad,fe,decode"
	req["request"]["result_type"] = "full"
	req["request"]["sequence"] = 1
	req["audio"] = make(map[string]interface{})
	req["audio"]["format"] = z.Cfg.format
	req["audio"]["codec"] = "raw"
	req["audio"]["rate"] = z.Cfg.sampleRate
	req["audio"]["language"] = z.Cfg.language
	reqStr, _ := json.Marshal(req)
	return reqStr
}

func (z *ZjAsr) parseResponse(msg []byte) (AsrResponse, error) {
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

	switch messageType {
	case byte(serverFullResponse):
		payloadSize = int(int32(binary.BigEndian.Uint32(payload[0:4])))
		payloadMsg = payload[4:]
	case byte(serverAck):
		seq := int32(binary.BigEndian.Uint32(payload[:4]))
		if len(payload) >= 8 {
			payloadSize = int(binary.BigEndian.Uint32(payload[4:8]))
			payloadMsg = payload[8:]
		}
		log.Infof("serverAck seq: %s", seq)
	case byte(serverErrorResponse):
		code := int32(binary.BigEndian.Uint32(payload[:4]))
		payloadSize = int(binary.BigEndian.Uint32(payload[4:8]))
		payloadMsg = payload[8:]
		log.Infof("serverErrorResponse code: %d", code)
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
		if err := json.Unmarshal(payloadMsg, &asrResponse); err != nil {
			return AsrResponse{}, err
		}
	}

	if asrResponse.Code != 1000 {
		return AsrResponse{}, fmt.Errorf("%s", asrResponse.Message)
	}

	return asrResponse, nil
}
