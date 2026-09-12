package util

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"
)

// GenerateVoucherNo 生成报名凭证号：GB + 日期 + 4 位随机数字。
func GenerateVoucherNo() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(10000))
	return fmt.Sprintf("GB%s%04d", time.Now().Format("20060102"), n.Int64())
}

// ValidateVoucherFormat 校验凭证号格式。
func ValidateVoucherFormat(voucher string) bool {
	voucher = strings.TrimSpace(voucher)
	if len(voucher) < 8 {
		return false
	}
	return strings.HasPrefix(voucher, "GB")
}
