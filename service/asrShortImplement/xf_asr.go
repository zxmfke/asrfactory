package asrShortImplement

import (
	"asrer/define"
	"asrer/log"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	statusFirstFrame    = 0
	statusContinueFrame = 1
	statusLastFrame     = 2
)

type XfAsr struct {
	Cfg    *XfAsrConfig
	TaskID string
}

type XfAsrConfig struct {
	AppID     string
	AppKey    string
	AppSecret string
}

func NewXfAsr(taskID string, cfg *XfAsrConfig) *XfAsr {
	return &XfAsr{Cfg: cfg, TaskID: taskID}
}

func (x *XfAsr) Do(fileName string, sampleRate define.AudioSampleRate) (string, error) {
	//st := time.Now()
	d := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}
	//握手并建立websocket 连接
	conn, resp, err := d.Dial(assembleAuthUrl("wss://iat-api.xfyun.cn/v2/iat", x.Cfg.AppKey, x.Cfg.AppSecret), nil)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 101 {
		return "", fmt.Errorf("识别失败，HTTP StatusCode: %d", resp.StatusCode)
	}

	//打开音频文件

	var frameSize = 1280                 //每一帧的音频大小
	var interval = 40 * time.Millisecond //发送音频间隔
	//开启协程，发送数据
	ctx, _ := context.WithCancel(context.Background())
	defer func() {
		_ = conn.Close()
	}()
	var status = 0
	go func() {
		//	start:
		audioFile, err := os.Open(fileName)
		if err != nil {
			panic(err)
		}
		status = statusFirstFrame //音频的状态信息，标识音频是第一帧，还是中间帧、最后一帧
		//		time.Sleep(20*time.Second)
		var buffer = make([]byte, frameSize)
		for {
			audioLen, err := audioFile.Read(buffer)
			if err != nil {
				if err == io.EOF { //文件读取完了，改变status = statusLastFrame
					status = statusLastFrame
				} else {
					panic(err)
				}
			}
			select {
			case <-ctx.Done():
				log.Info("session end ---")
				return
			default:
			}
			switch status {
			case statusFirstFrame: //发送第一帧音频，带business 参数
				frameData := map[string]interface{}{
					"common": map[string]interface{}{
						"app_id": x.Cfg.AppID, //appid 必须带上，只需第一帧发送
					},
					"business": map[string]interface{}{ //business 参数，只需一帧发送
						"language": "zh_cn",
						"domain":   "iat",
						"accent":   "mandarin",
					},
					"data": map[string]interface{}{
						"status":   statusFirstFrame,
						"format":   fmt.Sprintf("audio/L16;rate=%d", sampleRate),
						"audio":    base64.StdEncoding.EncodeToString(buffer[:audioLen]),
						"encoding": "raw",
					},
				}
				_ = conn.WriteJSON(frameData)
				status = statusContinueFrame
			case statusContinueFrame:
				frameData := map[string]interface{}{
					"data": map[string]interface{}{
						"status":   statusContinueFrame,
						"format":   fmt.Sprintf("audio/L16;rate=%d", sampleRate),
						"audio":    base64.StdEncoding.EncodeToString(buffer[:audioLen]),
						"encoding": "raw",
					},
				}
				_ = conn.WriteJSON(frameData)
			case statusLastFrame:
				frameData := map[string]interface{}{
					"data": map[string]interface{}{
						"status":   statusLastFrame,
						"format":   fmt.Sprintf("audio/L16;rate=%d", sampleRate),
						"audio":    base64.StdEncoding.EncodeToString(buffer[:audioLen]),
						"encoding": "raw",
					},
				}
				_ = conn.WriteJSON(frameData)
				return
				//	goto start
			}

			//模拟音频采样间隔
			time.Sleep(interval)
		}

	}()

	//获取返回的数据
	var decoder Decoder

	for {
		var respFromRead = respData{}
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Errorf("read message error: %s", err)
			break
		}
		_ = json.Unmarshal(msg, &respFromRead)
		//log.Infof(resp.Data.Result.String(), resp.Sid)
		if respFromRead.Code != 0 {
			//log.Infof(resp.Code, resp.Message, time.Since(st))
			return "", fmt.Errorf("%s", respFromRead.Message)
		}
		decoder.Decode(&respFromRead.Data.Result)
		if respFromRead.Data.Status == 2 {
			//log.Infof(resp.Code, resp.Message, time.Since(st))
			return decoder.String(), nil
		}

	}

	time.Sleep(1 * time.Second)
	return "", nil
}

type respData struct {
	Sid     string `json:"sid"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    data   `json:"data"`
}

type data struct {
	Result xfResult `json:"result"`
	Status int      `json:"status"`
}

// 创建鉴权url  apikey 即 hmac username
func assembleAuthUrl(hostURL string, apiKey, apiSecret string) string {
	ul, _ := url.Parse(hostURL)
	//签名时间
	date := time.Now().UTC().Format(time.RFC1123)
	//date = "Tue, 28 May 2019 09:10:42 MST"
	//参与签名的字段 host ,date, request-line
	signString := []string{"host: " + ul.Host, "date: " + date, "GET " + ul.Path + " HTTP/1.1"}
	//拼接签名字符串
	sgin := strings.Join(signString, "\n")
	//签名结果
	sha := hmacWithShaToBase64(sgin, apiSecret)
	//构建请求参数 此时不需要urlencoding
	authUrl := fmt.Sprintf("hmac username=\"%s\", algorithm=\"%s\", headers=\"%s\", signature=\"%s\"", apiKey,
		"hmac-sha256", "host date request-line", sha)
	//将请求参数使用base64编码
	authorization := base64.StdEncoding.EncodeToString([]byte(authUrl))

	v := url.Values{}
	v.Add("host", ul.Host)
	v.Add("date", date)
	v.Add("authorization", authorization)
	//将编码后的字符串url encode后添加到url后面
	return hostURL + "?" + v.Encode()
}

func hmacWithShaToBase64(data, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(data))
	encodeData := mac.Sum(nil)
	return base64.StdEncoding.EncodeToString(encodeData)
}

// 解析返回数据，仅供demo参考，实际场景可能与此不同。
type Decoder struct {
	results []*xfResult
}

func (d *Decoder) Decode(deepResult *xfResult) {
	if len(d.results) <= deepResult.Sn {
		d.results = append(d.results, make([]*xfResult, deepResult.Sn-len(d.results)+1)...)
	}
	if deepResult.Pgs == "rpl" {
		for i := deepResult.Rg[0]; i <= deepResult.Rg[1]; i++ {
			d.results[i] = nil
		}
	}
	d.results[deepResult.Sn] = deepResult
}

func (d *Decoder) String() string {
	var r string
	for _, v := range d.results {
		if v == nil {
			continue
		}
		r += v.String()
	}
	return r
}

type xfResult struct {
	Ls  bool   `json:"ls"`
	Rg  []int  `json:"rg"`
	Sn  int    `json:"sn"`
	Pgs string `json:"pgs"`
	Ws  []ws   `json:"ws"`
}

func (t *xfResult) String() string {
	var wss string
	for _, v := range t.Ws {
		wss += v.String()
	}
	return wss
}

type ws struct {
	Bg int  `json:"bg"`
	Cw []cw `json:"cw"`
}

func (w *ws) String() string {
	var wss string
	for _, v := range w.Cw {
		wss += v.W
	}
	return wss
}

type cw struct {
	Sc int    `json:"sc"`
	W  string `json:"w"`
}
