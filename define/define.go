package define

import (
	"fmt"
)

type AudioSampleRate = int

const (
	SampleRate16k AudioSampleRate = 16000
	SampleRate8k  AudioSampleRate = 8000
)

type Output struct {
	Status OutputStatus `json:"status"`
	Text   string       `json:"text"`
	Err    error        `json:"err"`
}

func (o Output) String() string {
	return fmt.Sprintf("status : %s | text : %s", o.Status, o.Text)
}

func (o Output) IsDone() bool {
	return o.Status == Done
}

type OutputStatus string

const (
	Start           OutputStatus = "start"
	SentencePartial OutputStatus = "partial"
	SentenceFinal   OutputStatus = "final"
	Done            OutputStatus = "done"
)

type AsrType = int

const (
	_ AsrType = iota
	AsrKst
	AsrAli
	AsrXf
	AsrTx
	AsrBd
	AsrZj
)
