package util

import "fmt"

// SscanUint64 解析二维码内容中的报名 ID。
func SscanUint64(s string) (uint64, error) {
	var v uint64
	if _, err := fmt.Sscanf(s, "%d", &v); err != nil {
		return 0, err
	}
	return v, nil
}
