package utils

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"os"
	"time"
)

// GetProjectDir 获取项目根目录
func GetProjectDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	return dir
}

// 辅助函数：返回两个时间间隔中较小的一个
func MinDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

// GenerateRandomString 生成指定长度的随机字符串
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateRandomPassword 生成指定长度的随机密码
func GenerateRandomPassword(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+"
	password := make([]byte, length)
	for i := range password {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			// 在密码生成中，如果出现错误，可以panic，因为这通常是系统熵源问题
			panic(err)
		}
		password[i] = chars[num.Int64()]
	}
	return string(password)
}
