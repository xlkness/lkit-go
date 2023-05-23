package utils

import (
	"bytes"
	"crypto/rand"
	"math/big"
	"strings"
)

var str = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

// GetGlobalIDName
func GetGlobalIDName(serviceName string) (string, string) {
	if serviceName == "" {
		return randomString(7), serviceName
	}

	// 查看名字是否带-，注入的名字是'name-id'，例如k8s环境
	strs := strings.Split(serviceName, "-")
	if len(strs) == 1 {
		return randomString(7), serviceName
	}

	return strs[len(strs)-1], strs[0]
}

// GetGlobalIDFromRandomString 随机一个全局id
func GetGlobalIDFromRandomString(len int) string {
	return randomString(len)
}

// randomString 随机任意长度字符串，长度6大概5w次左右有概率重复
func randomString(len int) string {
	var container string
	b := bytes.NewBufferString(str)
	length := b.Len()
	bigInt := big.NewInt(int64(length))
	for i := 0; i < len; i++ {
		randomInt, _ := rand.Int(rand.Reader, bigInt)
		container += string(str[randomInt.Int64()])
	}
	return container
}
