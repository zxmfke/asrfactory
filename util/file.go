package util

import (
	"asrer/log"
	"os"
	"time"
)

// ReadPerSize 每 size 大小读取文件内容
func ReadPerSize(filePath string, size int, do func([]byte) error) error {

	var (
		buf = make([]byte, size)
	)

	audio, err := os.Open(filePath)
	defer func() {
		_ = audio.Close()
	}()
	if err != nil {
		log.Errorf("打开文件失败: %s", err.Error())
		return err
	}

	for i, e := audio.Read(buf); i > 0; i, e = audio.Read(buf) {

		if e != nil {
			if e.Error() == "EOF" {
				break
			}
			log.Errorf("读取文件错误: %s", err.Error())
			return err
		}

		if err = do(buf[:i]); err != nil {
			log.Errorf("%s", err.Error())
			return err
		}

		time.Sleep(40 * time.Millisecond)
	}

	log.Info("send finish")

	return nil
}
