package asrShort

import (
	"asrer/define"
	"asrer/service/asrShortImplement"
)

type Asr interface {
	Do(fileName string, sampleRate define.AudioSampleRate) (string, error)
}

func NewAsr(taskID string, typ define.AsrType) Asr {
	switch typ {
	case define.AsrKst:
		return asrShortImplement.NewKstAsr(taskID, &asrShortImplement.KstAsrConfig{})
	case define.AsrAli:
		return asrShortImplement.NewAliAsr(taskID, &asrShortImplement.AliAsrConfig{
			Region:  "cn-shanghai",
			Domain:  "nls-meta.cn-shanghai.aliyuncs.com",
			Version: "2019-02-28",
		})
	case define.AsrXf:
		return asrShortImplement.NewXfAsr(taskID, &asrShortImplement.XfAsrConfig{})
	case define.AsrTx:
		return asrShortImplement.NewTxAsr(taskID, &asrShortImplement.TxAsrConfig{})
	case define.AsrBd:
		return asrShortImplement.NewBdAsr(taskID, &asrShortImplement.BdAsrConfig{})
	case define.AsrZj:
		return asrShortImplement.NewZjAsr(taskID, &asrShortImplement.ZjAsrConfig{})
	}

	return nil

}
