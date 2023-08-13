package asrShortImplement

import (
	"asrer/ahttp"
	"asrer/define"
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type KstAsr struct {
	Cfg    *KstAsrConfig
	TaskID string
}

type KstAsrConfig struct {
	AppID     string
	AppSecret string
}

type kstTokenResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data tR     `json:"data"`
}

type tR struct {
	AccessToken string `json:"access_token"`
}

func NewKstAsr(taskID string, cfg *KstAsrConfig) *KstAsr {
	return &KstAsr{Cfg: cfg, TaskID: taskID}
}

func (k *KstAsr) getToken() (string, error) {

	var (
		err       error
		url       = "https://aihc.shengwenyun.com/aihc/auth"
		payload   = strings.NewReader(`{"app_id": "` + k.Cfg.AppID + `","app_secret": "` + k.Cfg.AppSecret + `"}`)
		tokenResp = new(kstTokenResp)
	)

	if err = ahttp.Do(url, "POST", []ahttp.HeaderKV{{Key: "Content-Type", Value: "application/json"}}, payload, tokenResp); err != nil {
		return "", err
	}

	if tokenResp.Code != 0 {
		return "", fmt.Errorf("%s", tokenResp.Msg)
	}

	return tokenResp.Data.AccessToken, nil
}

func (k *KstAsr) Do(fileName string, sampleRate define.AudioSampleRate) (string, error) {

	var (
		err     error
		url     = "https://aihc.shengwenyun.com/aihc/v1/asr/api"
		resp    = new(kstAsrResp)
		payload = &bytes.Buffer{}
		writer  = multipart.NewWriter(payload)
		token   string
	)

	file, err := os.Open(fileName)
	if err != nil {
		return "", err
	}

	defer func() {
		_ = file.Close()
	}()

	part1, err := writer.CreateFormFile("file", filepath.Base(fileName))
	if err != nil {
		return "", err
	}

	if _, err = io.Copy(part1, file); err != nil {
		return "", err
	}
	_ = writer.WriteField("sample_rate", strconv.Itoa(sampleRate))
	_ = writer.WriteField("wave_format", filepath.Ext(fileName)[1:])
	_ = writer.Close()

	if token, err = k.getToken(); err != nil {
		return "", err
	}

	if err = ahttp.Do(url, "POST", []ahttp.HeaderKV{
		{Key: "Authorization", Value: "Bearer " + token},
		{Key: "Content-Type", Value: writer.FormDataContentType()},
	}, payload, resp); err != nil {
		return "", err
	}

	if resp.Code != 0 {
		return "", fmt.Errorf("%s", resp.Msg)
	}

	return resp.Data.Result, nil
}

type kstAsrResp struct {
	Code int
	Msg  string
	Data struct {
		Result string `json:"result"`
	} `json:"data"`
}
