package asrer

import (
	"asrer/asrStreamer"
	"asrer/define"
	"asrer/log"
	"asrer/service/asrStreamerImplement"
	"asrer/util"
	"fmt"
	"testing"
	"time"
)

type AsrRun struct {
	sampleRate define.AudioSampleRate
	format     string
}

var asrRun = map[int]AsrRun{
	define.AsrAli: {
		sampleRate: define.SampleRate16k,
		format:     "wav",
	},
	define.AsrBd: {
		sampleRate: define.SampleRate16k,
		format:     "pcm",
	},
	define.AsrKst: {
		sampleRate: define.SampleRate16k,
		format:     "wav",
	},
	define.AsrTx: {
		sampleRate: define.SampleRate16k,
		format:     "wav",
	},
	define.AsrXf: {
		sampleRate: define.SampleRate16k,
		format:     "pcm",
	},
	define.AsrZj: {
		sampleRate: define.SampleRate16k,
		format:     "wav",
	},
}

func TestAsrStream(t *testing.T) {
	var (
		svc    = define.AsrAli
		taskID = util.GetUuid()
		err    error
	)

	asrWs, asrOutput := asrStreamer.NewAsrStreamer(taskID, asrStreamerImplement.NewAsrWs(taskID, svc, asrRun[svc].sampleRate))

	if err = asrWs.Init(); err != nil {
		t.Error(err.Error())
		return
	}
	var (
		closeSign = make(chan struct{})
		startSign = make(chan error)
	)
	go AsrOutputGo(asrOutput, startSign, closeSign)

	time.Sleep(time.Second * 1)

	// 判断是否能开始发送数据
	startErr := <-startSign
	if startErr != nil {
		log.Errorf("start err : %s", startErr.Error())
		close(startSign)
		close(closeSign)
		return
	}

	var (
		filePath = fmt.Sprintf("./testdata/test1_%d.%s", asrRun[svc].sampleRate, asrRun[svc].format)
		size     = 3200 // 发送数据的大小
	)

	if err = util.ReadPerSize(filePath, size, asrWs.Send); err != nil {
		log.Errorf("asr ws send fail : %s", err.Error())
		return
	}

	asrWs.End()

	select {
	case <-closeSign:
		close(closeSign)
		log.Info("finished")
	}
}

func AsrOutputGo(output <-chan define.Output, startSign chan error, closeSign chan struct{}) {
	for {
		select {
		case ot := <-output:
			if ot.Err != nil {
				log.Errorf("output err : %s", ot.Err)

				if ot.Status == define.Start {
					startSign <- ot.Err
				}
				return
			}

			if ot.Status == define.SentenceFinal {
				log.Infof("输出结果: %s", ot.String())
			}

			if ot.Status == define.Start {
				startSign <- nil
				break
			}

			if ot.IsDone() {
				closeSign <- struct{}{}
				return
			}
		}
	}
}
