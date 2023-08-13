package kstAsrStreamer

import "asrer/define"

type KstAsrConfig struct {
	AppID      string
	AppSecret  string
	SampleRate define.AudioSampleRate
}

func (k KstAsrConfig) url() string {

	wssURL := "wss://aihc.shengwenyun.com/aihc/v1/asr/stream"

	if k.SampleRate == define.SampleRate8k {
		wssURL += "/8k"
	}

	return wssURL
}

type Resp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data tR     `json:"data"`
}

type tR struct {
	AccessToken string `json:"access_token"`
}

type WsResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data Data   `json:"data"`
}

type Data struct {
	Result string       `json:"result"`
	Ss     int          `json:"ss"`
	Se     int          `json:"se"`
	Wp     []*WordPiece `json:"wp"`
	Typ    int          `json:"type"`
}

type WordPiece struct {
	Word  string `json:"word"`
	Start int    `json:"start"`
	End   int    `json:"end"`
}

type WSMode struct {
	Mode string `json:"mode"`
}
