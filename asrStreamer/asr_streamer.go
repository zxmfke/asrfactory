package asrStreamer

import (
	"asrer/define"
	"asrer/log"
	"asrer/service/asrStreamerImplement"
	"fmt"
)

type AsrStreamer struct {
	TaskID           string               // task id
	close            bool                 // 是否关闭
	init             bool                 // 是否初始化
	AsrWsOutputChan  chan define.Output   // asr stream 返回的识别结果
	callerOutputChan <-chan define.Output // 给调用者接受数据的 channel
	otChan           chan define.Output   // 内部 channel 传数据，避免 close
	asrWs            asrStreamerImplement.AsrWs
}

// NewAsrStreamer new 一个 asr streamer
func NewAsrStreamer(taskID string, asrWs asrStreamerImplement.AsrWs) (*AsrStreamer, <-chan define.Output) {

	otChan := make(chan define.Output)
	outputChan := make(<-chan define.Output)
	outputChan = otChan

	asrStreamer := &AsrStreamer{
		TaskID:           taskID,
		close:            false,
		AsrWsOutputChan:  make(chan define.Output),
		callerOutputChan: outputChan,
		otChan:           otChan,
		asrWs:            asrWs,
	}

	return asrStreamer, asrStreamer.callerOutputChan
}

// Init 初始化
func (a *AsrStreamer) Init() error {

	if a.asrWs == nil {
		return fmt.Errorf("asr ws is nil")
	}

	if a.init {
		return fmt.Errorf("already init")
	}

	if err := a.asrWs.Init(); err != nil {
		return err
	}

	go a.live()
	go a.asrWs.Recv(a.AsrWsOutputChan)

	a.init = true

	return nil
}

func (a *AsrStreamer) live() {
	for {
		select {
		case output := <-a.AsrWsOutputChan:
			if output.Err != nil {

				a.otChan <- output
				a.Close()

				log.Errorf("output err : %s", output.Err.Error())
				return
			}

			if !a.close {
				a.otChan <- output
			}

			if output.Status == define.Done {
				a.Close()
				return
			}
		}
	}
}

// Send 发送实时语音数据
func (a *AsrStreamer) Send(data []byte) error {
	if a.close || !a.init {
		return nil
	}

	return a.asrWs.Send(data)
}

// End asr 实时流数据发送结束
func (a *AsrStreamer) End() {
	if a.close || !a.init {
		return
	}

	_ = a.asrWs.End()
}

// Close 主动关闭
func (a *AsrStreamer) Close() {
	if a.close {
		return
	}

	a.asrWs.Close()

	a.close = true
	close(a.otChan)
	close(a.AsrWsOutputChan)
}
