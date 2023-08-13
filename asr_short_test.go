package asrer

import (
	"asrer/asrShort"
	"asrer/define"
	"fmt"
	"testing"
	"time"
)

func TestAsrShort(t *testing.T) {

	cases := []struct {
		name       string
		sampleRate define.AudioSampleRate
		ext        string
		asrType    define.AsrType
	}{
		{
			name:       "阿里测试",
			asrType:    define.AsrAli,
			sampleRate: define.SampleRate16k,
			ext:        "wav",
		}, {
			name:       "快商通测试",
			asrType:    define.AsrKst,
			sampleRate: define.SampleRate16k,
			ext:        "wav",
		}, {
			name:       "腾讯测试",
			asrType:    define.AsrTx,
			sampleRate: define.SampleRate16k,
			ext:        "wav",
		}, {
			name:       "百度测试",
			asrType:    define.AsrBd,
			sampleRate: define.SampleRate16k,
			ext:        "pcm",
		}, {
			name:       "讯飞测试",
			asrType:    define.AsrXf,
			sampleRate: define.SampleRate16k,
			ext:        "pcm",
		}, {
			name:       "字节测试",
			asrType:    define.AsrZj,
			sampleRate: define.SampleRate16k,
			ext:        "wav",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			asrOperator := asrShort.NewAsr(fmt.Sprintf("%d", time.Now().Unix()), define.AsrAli)

			result, err := asrOperator.Do(fmt.Sprintf("./testdata/test1_%d.%s", tt.sampleRate, tt.ext), tt.sampleRate)
			if err != nil {
				t.Errorf("%s", err.Error())
				return
			}

			t.Logf("%s", result)
		})
	}
}
