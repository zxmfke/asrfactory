package bdAsrStreamer

import "asrer/define"

type BdAsrConfig struct {
	AppID      int
	AppKey     string
	SampleRate define.AudioSampleRate
}

type bdAsrReq struct {
	Type string    `json:"type"`
	Data bdReqData `json:"data"`
}

type bdReqData struct {
	AppID  int    `json:"appid"`
	AppKey string `json:"appkey"`
	DevPID int    `json:"dev_pid"` // 识别模型，比如普通话还是英语，是否要加标点等
	LmID   int    `json:"lm_id"`   // 自训练平台才有这个参数
	Cuid   string `json:"cuid"`    // 随便填不影响使用。机器的mac或者其它唯一id，页面上计算UV用。
	Format string `json:"format"`
	Sample int    `json:"sample"`
}

type bdAsrResult struct {
	ErrNo     int    `json:"err_no"`
	ErrMsg    string `json:"err_msg"`
	Type      string `json:"type"`
	Result    string `json:"result"`
	StartTime int    `json:"start_time"`
	EndTime   int    `json:"end_time"`
	LogId     int    `json:"log_id"`
	Sn        string `json:"sn"`
}

const (
	Start  = "START"
	Finish = "FINISH"
	Cancel = "CANCEL"

	MidText   = "MID_TEXT"
	FinalText = "FIN_TEXT"
	HeartBeat = "HEARTBEAT"
)
