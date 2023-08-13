package asrShortImplement

import (
	"asrer/ahttp"
	"asrer/define"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// https://cloud.baidu.com/doc/SPEECH/s/Jlbxdezuf

type bdReqData struct {
	Format  string      `json:"format"`
	Rate    int         `json:"rate"`
	DevPid  int         `json:"dev_pid"`
	Channel int         `json:"channel"`
	Token   interface{} `json:"token"`
	Cuid    string      `json:"cuid"`
	Len     int         `json:"len"`
	Speech  string      `json:"speech"`
}

type BdAsr struct {
	Cfg    *BdAsrConfig
	TaskID string
}

type BdAsrConfig struct {
	AppID     int
	AppKey    string
	AppSecret string
}

type dbAsrReq struct {
	Format  string `json:"format"`
	Rate    string `json:"rate"`
	Channel int    `json:"channel"`
	Speech  string `json:"speech"`
	Len     int    `json:"len"`
	Cuid    string `json:"cuid"`
	DevPid  int    `json:"dev_pid"`
	Token   string `json:"token"`
}

func NewBdAsr(taskID string, cfg *BdAsrConfig) *BdAsr {
	return &BdAsr{Cfg: cfg, TaskID: taskID}
}

func (b *BdAsr) Do(fileName string, sampleRate define.AudioSampleRate) (string, error) {

	var (
		err   error
		fdata []byte
		token string
	)

	if fdata, err = os.ReadFile(fileName); err != nil {
		return "", err
	}

	if token, err = b.getToken(); err != nil {
		return "", err
	}

	var (
		url    = "https://vop.baidu.com/server_api"
		asrReq = dbAsrReq{
			Format:  filepath.Ext(fileName)[1:],
			Rate:    strconv.Itoa(sampleRate),
			Channel: 1,
			Speech:  base64.StdEncoding.EncodeToString(fdata),
			Len:     len(fdata),
			Cuid:    b.TaskID,
			DevPid:  1537,
			Token:   token,
		}
		asrReqJson, _ = json.Marshal(asrReq)
		payload       = strings.NewReader(string(asrReqJson))
		header        = []ahttp.HeaderKV{
			{
				Key:   "Content-Type",
				Value: "application/json",
			},
			{
				Key:   "Accept",
				Value: "application/json",
			},
		}
		resp = new(BdAsrResp)
	)

	if err = ahttp.Do(url, "POST", header, payload, &resp); err != nil {
		return "", err
	}

	if resp.ErrNo != 0 {
		return "", fmt.Errorf("%s", resp.ErrMsg)
	}

	return resp.String(), nil
}

func (b *BdAsr) getToken() (string, error) {

	var (
		err      error
		url      = "https://aip.baidubce.com/oauth/2.0/token"
		postData = fmt.Sprintf("grant_type=client_credentials&client_id=%s&client_secret=%s", b.Cfg.AppKey, b.Cfg.AppSecret)
		header   = []ahttp.HeaderKV{
			{
				Key:   "Content-Type",
				Value: "application/x-www-form-urlencoded",
			},
		}
		resp = new(bdTokenRes)
	)

	if err = ahttp.Do(url, "POST", header, strings.NewReader(postData), resp); err != nil {
		return "", err
	}

	return resp.AccessToken, nil
}

type BdAsrResp struct {
	ErrNo    int      `json:"err_no"`
	ErrMsg   string   `json:"err_msg"`
	CorpusNo string   `json:"corpus_no"`
	Sn       string   `json:"sn"`
	Result   []string `json:"result"`
}

func (b *BdAsrResp) String() string {
	var result = ""

	for i := 0; i < len(b.Result); i++ {
		result += b.Result[i]
	}

	return result
}

type bdTokenRes struct {
	RefreshToken  string `json:"refresh_token"`
	ExpiresIn     int    `json:"expires_in"`
	SessionKey    string `json:"session_key"`
	AccessToken   string `json:"access_token"`
	Scope         string `json:"scope"`
	SessionSecret string `json:"session_secret"`
}
