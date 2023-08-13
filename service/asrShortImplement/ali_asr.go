package asrShortImplement

import (
	"asrer/ahttp"
	"asrer/define"
	"bytes"
	"encoding/json"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
	"os"
	"path/filepath"
	"strconv"
)

// https://help.aliyun.com/document_detail/92131.html?spm=a2c4g.432038.0.0.533f74cbWU1MuL#section-og9-qpl-2jq

type AliAsr struct {
	Cfg    *AliAsrConfig
	TaskID string
}

type AliAsrConfig struct {
	AppID     string
	AppKey    string
	AppSecret string
	Region    string
	Domain    string
	Version   string
}

func NewAliAsr(taskID string, cfg *AliAsrConfig) *AliAsr {
	return &AliAsr{Cfg: cfg, TaskID: taskID}
}

func (a *AliAsr) Do(fileName string, sampleRate define.AudioSampleRate) (string, error) {

	var (
		err       error
		token     string
		ext       = filepath.Ext(fileName)
		audioData []byte
	)
	var url = "https://nls-gateway-cn-shanghai.aliyuncs.com/stream/v1/asr"
	url = url + "?appkey=" + a.Cfg.AppKey
	url = url + "&format=" + ext[1:]
	url = url + "&sample_rate=" + strconv.Itoa(sampleRate)
	url = url + "&enable_punctuation_prediction=true"

	if audioData, err = os.ReadFile(fileName); err != nil {
		return "", err
	}

	var (
		resultMap = make(map[string]interface{})
	)

	if token, err = a.getToken(); err != nil {
		return "", err
	}

	header := []ahttp.HeaderKV{
		{
			Key:   "X-NLS-Token",
			Value: token,
		},
		{
			Key:   "Content-Type",
			Value: "application/octet-stream",
		},
	}

	if err = ahttp.Do(url, "POST", header, bytes.NewBuffer(audioData), &resultMap); err != nil {
		return "", err
	}

	return resultMap["result"].(string), nil
}

func (a *AliAsr) getToken() (string, error) {
	var (
		err      error
		revMsg   = new(AliTokenRes)
		client   *sdk.Client
		response *responses.CommonResponse
	)

	if client, err = sdk.NewClientWithAccessKey(a.Cfg.Region, a.Cfg.AppID, a.Cfg.AppSecret); err != nil {
		return "", err
	}

	request := requests.NewCommonRequest()
	request.Method = "POST"
	request.Domain = a.Cfg.Domain
	request.ApiName = "CreateToken"
	request.Version = a.Cfg.Version

	if response, err = client.ProcessCommonRequest(request); err != nil {
		return "", err
	}

	if err := json.Unmarshal(response.GetHttpContentBytes(), revMsg); err != nil {
		return "", err
	}

	return revMsg.Token.Id, err
}

type AliTokenRes struct {
	ErrMsg string   `json:"ErrMsg"`
	Token  aliToken `json:"Token"`
}

type aliToken struct {
	UserId     string `json:"UserId"`
	Id         string `json:"Id"`
	ExpireTime int64  `json:"ExpireTime"`
}
