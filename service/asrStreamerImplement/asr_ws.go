package asrStreamerImplement

import (
	"asrer/define"
	"asrer/service/asrStreamerImplement/aliAsrStreamer"
	"asrer/service/asrStreamerImplement/bdAsrStreamer"
	"asrer/service/asrStreamerImplement/kstAsrStreamer"
	"asrer/service/asrStreamerImplement/txAsrStreamer"
	"asrer/service/asrStreamerImplement/xfAsrStreamer"
	"asrer/service/asrStreamerImplement/zjAsrStreamer"
)

type AsrWs interface {
	Init() error

	Send([]byte) error

	Recv(chan<- define.Output)

	End() error

	Close()
}

func NewAsrWs(taskID string, typ define.AsrType, sampleRate define.AudioSampleRate) AsrWs {

	switch typ {
	case define.AsrKst:
		return kstAsrStreamer.NewKstAsrStreamer(taskID, kstAsrStreamer.KstAsrConfig{
			SampleRate: sampleRate,
		})
	case define.AsrAli:
		return aliAsrStreamer.NewAliAsrStreamer(taskID, aliAsrStreamer.AliAsrConfig{
			Region:     "cn-shanghai",
			Domain:     "nls-meta.cn-shanghai.aliyuncs.com",
			Version:    "2019-02-28",
			SampleRate: sampleRate,
		})
	case define.AsrXf:
		return xfAsrStreamer.NewXfAsrStreamer(taskID, xfAsrStreamer.XfAsrConfig{
			SampleRate: sampleRate,
		})
	case define.AsrTx:
		return txAsrStreamer.NewTxAsrStreamer(taskID, txAsrStreamer.TxAsrConfig{
			SampleRate: sampleRate,
		})
	case define.AsrBd:
		return bdAsrStreamer.NewBdAsrStreamer(taskID, bdAsrStreamer.BdAsrConfig{
			SampleRate: sampleRate,
		})

	case define.AsrZj:
		return zjAsrStreamer.NewZjAsrStreamer(taskID, zjAsrStreamer.ZjAsrConfig{
			Format:     "wav",
			SampleRate: sampleRate,
			Language:   "zh-CN",
		})
	}

	return nil
}
