package util

import (
	"strings"
)

func GetUuid() string {
	s := NewV1()
	return strings.ReplaceAll(s.String(), "-", "")
}
