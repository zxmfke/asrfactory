package ahttp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type HeaderKV struct {
	Key   string
	Value string
}

func Do(url, method string, header []HeaderKV, payload io.Reader, target interface{}) error {

	var (
		err      error
		request  *http.Request
		response *http.Response
		client   = &http.Client{}
	)

	if request, err = http.NewRequest(method, url, payload); err != nil {
		return err
	}

	for _, kv := range header {
		request.Header.Add(kv.Key, kv.Value)
	}

	if response, err = client.Do(request); err != nil {
		return err
	}

	defer func() {
		_ = response.Body.Close()
	}()

	body, _ := io.ReadAll(response.Body)

	if response.StatusCode != 200 {
		return fmt.Errorf("识别失败，HTTP StatusCode: %d", response.StatusCode)
	}

	if err = json.Unmarshal(body, target); err != nil {
		return err
	}

	return nil
}
