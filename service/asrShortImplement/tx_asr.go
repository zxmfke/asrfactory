package asrShortImplement

import (
	"asrer/define"
	"encoding/base64"
	errors2 "errors"
	asr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/asr/v20190614"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"os"
	"path/filepath"
)

// https://cloud.tencent.cn/document/api/1093/35646

type TxAsrConfig struct {
	AppID     string
	AppKey    string
	AppSecret string
}

type TxAsr struct {
	Cfg    *TxAsrConfig
	TaskID string
}

type txAsrResp struct {
	Response struct {
		RequestId     string `json:"RequestId"`
		Result        string `json:"result"`
		AudioDuration int    `json:"AudioDuration"`
		WordSize      int    `json:"WordSize"`
		WordList      []struct {
			Word      string `json:"Word"`
			StartTime int    `json:"StartTime"`
			EndTime   int    `json:"EndTime"`
		} `json:"WordList"`
	} `json:"Response"`
}

func NewTxAsr(taskID string, cfg *TxAsrConfig) *TxAsr {
	return &TxAsr{Cfg: cfg, TaskID: taskID}
}

func (x *TxAsr) Do(fileName string, sampleRate define.AudioSampleRate) (string, error) {

	var (
		err   error
		fdata []byte
		sp    = "8k_zh"
	)

	if sampleRate == define.SampleRate16k {
		sp = "16k_zh"
	}

	if fdata, err = os.ReadFile(fileName); err != nil {
		return "", err
	}

	credential := common.NewCredential(
		x.Cfg.AppSecret,
		x.Cfg.AppKey,
	)

	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "asr.tencentcloudapi.com"

	client, _ := asr.NewClient(credential, "", cpf)

	request := asr.NewSentenceRecognitionRequest()

	request.EngSerViceType = common.StringPtr(sp)
	request.SourceType = common.Uint64Ptr(1)
	request.VoiceFormat = common.StringPtr(filepath.Ext(fileName)[1:])
	request.Data = common.StringPtr(base64.StdEncoding.EncodeToString(fdata))
	request.DataLen = common.Int64Ptr(int64(len(fdata)))

	response, err := client.SentenceRecognition(request)
	var tencentCloudSDKError *errors.TencentCloudSDKError
	if errors2.As(err, &tencentCloudSDKError) {
		return "", err
	}
	if err != nil {
		return "", err
	}

	// 输出json格式的字符串回包
	return *response.Response.Result, nil
}
