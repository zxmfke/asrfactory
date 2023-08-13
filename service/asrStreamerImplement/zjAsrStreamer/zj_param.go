package zjAsrStreamer

import (
	"asrer/define"
	"bytes"
	"compress/gzip"
	"io/ioutil"
)

type ZjAsrConfig struct {
	AppID      string
	Token      string
	Cluster    string
	workflow   string
	Format     string
	SampleRate define.AudioSampleRate
	Language   string
}

type ProtocolVersion byte
type MessageType byte
type MessageTypeSpecificFlags byte
type SerializationType byte
type CompressionType byte

const (
	SuccessCode = 1000

	SERVER_FULL_RESPONSE  = MessageType(0b1001)
	SERVER_ACK            = MessageType(0b1011)
	SERVER_ERROR_RESPONSE = MessageType(0b1111)

	JSON = SerializationType(0b0001)

	GZIP = CompressionType(0b0001)
)

var DefaultFullClientWsHeader = []byte{0x11, 0x10, 0x11, 0x00}
var DefaultAudioOnlyWsHeader = []byte{0x11, 0x20, 0x11, 0x00}
var DefaultLastAudioWsHeader = []byte{0x11, 0x22, 0x11, 0x00}

func gzipCompress(input []byte) []byte {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	w.Write(input)
	w.Close()
	return b.Bytes()
}

func gzipDecompress(input []byte) []byte {
	b := bytes.NewBuffer(input)
	r, _ := gzip.NewReader(b)
	out, _ := ioutil.ReadAll(r)
	r.Close()
	return out
}

type AsrResponse struct {
	Reqid    string     `json:"reqid"`
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
