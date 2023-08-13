package aliAsrStreamer

import "asrer/define"

const (
	AsrStreamStart    = "StartTranscription"
	AsrStreamStop     = "StopTranscription"
	start             = "TranscriptionStarted"
	sentenceBegin     = "SentenceBegin"
	sentenceOnChanged = "TranscriptionResultChanged"
	sentenceEnd       = "SentenceEnd"
	completed         = "TranscriptionCompleted"

	DailOK   = 101
	StatusOK = 20000000
)

type AliAsrConfig struct {
	AppID      string
	AppKey     string
	AppSecret  string
	Region     string
	Domain     string
	Version    string
	SampleRate define.AudioSampleRate
}

type AliTokenRes struct {
	ErrMsg string `json:"ErrMsg"`
	Token  struct {
		UserId     string `json:"UserId"`
		Id         string `json:"Id"`
		ExpireTime int64  `json:"ExpireTime"`
	} `json:"Token"`
}

type Header struct {
	MessageID  string `json:"message_id"`
	TaskID     string `json:"task_id"`
	Namespace  string `json:"namespace"`
	Name       string `json:"name"`
	AppKey     string `json:"appkey"`
	Status     int64  `json:"status"`
	StatusText string `json:"status_text"`
}

type CommonRequest struct {
	Header  Header      `json:"header"`
	Payload interface{} `json:"payload,omitempty"`
}

type TranscriptionResultChangedResp struct {
	Index  int    `json:"index"`
	Time   int    `json:"time"`
	Result string `json:"result"`
	Words  []Word `json:"words"`
}

type TranscriptionResultEndResp struct {
	Index     int    `json:"index"`
	Time      int    `json:"time"`
	Result    string `json:"result"`
	BeginTime int    `json:"begin_time"`
	Words     []Word `json:"words"`
}

type Word struct {
	Text      string `json:"text"`
	StartTime int    `json:"startTime"`
	EndTime   int    `json:"endTime"`
}

type StartTranscriptionParam struct {
	Format                         string `json:"format"`
	SampleRate                     int    `json:"sample_rate"`
	EnableIntermediateResult       bool   `json:"enable_intermediate_result"`        // 是否返回中间识别结果，默认是false。
	EnablePunctuationPrediction    bool   `json:"enable_punctuation_prediction"`     // 是否在后处理中添加标点，默认是false。
	EnableInverseTextNormalization bool   `json:"enable_inverse_text_normalization"` // ITN（逆文本inverse text normalization）中文数字转换阿拉伯数字。设置为True时，中文数字将转为阿拉伯数字输出，默认值：False。
	MaxSentenceSilence             int    `json:"max_sentence_silence"`              // 语音断句检测阈值，静音时长超过该阈值会被认为断句，参数范围200ms～2000ms，默认值800ms。
	EnableWords                    bool   `json:"enable_words"`                      // 是否开启返回词信息，默认是false。
}

func (a *AliAsrStreamer) defaultStartTranscriptionParam(wavTyp string) StartTranscriptionParam {
	return StartTranscriptionParam{
		Format:                         wavTyp,
		SampleRate:                     a.Cfg.SampleRate,
		EnableIntermediateResult:       true,
		EnablePunctuationPrediction:    true,
		EnableInverseTextNormalization: true,
		EnableWords:                    true,
		MaxSentenceSilence:             800,
	}
}
