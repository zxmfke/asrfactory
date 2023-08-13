package txAsrStreamer

import "asrer/define"

type TxAsrConfig struct {
	AppID       string
	AppKey      string
	AppSecretID string
	SampleRate  define.AudioSampleRate
}

type txAsrResp struct {
	Code      int             `json:"code"`
	Message   string          `json:"message"`
	VoiceID   string          `json:"voice_id"`
	MessageID string          `json:"message_id"`
	Result    txAsrRespResult `json:"result"`
	Final     int             `json:"final"`
}

type txAsrRespResult struct {
	SliceType    int    `json:"slice_type"`
	Index        int    `json:"index"`
	StartTime    int    `json:"start_time"`
	EndTime      int    `json:"end_time"`
	VoiceTextStr string `json:"voice_text_str"`
	WordSize     int    `json:"word_size"`
	WordList     []wl   `json:"word_list"`
}

type wl struct {
	Word       string `json:"word"`
	StartTime  int    `json:"start_time"`
	EndTime    int    `json:"end_time"`
	StableFlag int    `json:"stable_flag"`
}
