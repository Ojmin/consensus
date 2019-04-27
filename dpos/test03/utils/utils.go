package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

type UUID [16]byte

// 保留两位小数
func Decimal(value float64) float64 {
	value, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", value), 64)
	return value
}

// 加载配置信息
func LoadEnv(fileName string) {
	err := godotenv.Load(fileName)
	if err != nil {
		log.Fatal(err)
	}
}

// 根据key查找.env文件中的值
func GetEnvValue(key string) string {
	return os.Getenv(key)
}

// 计算哈希
func CalculateHash(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// 将string转成固定长度数据
func ConvertStrToBytes(str string) []byte {
	var dataBytes [constLength]byte
	for i, c := range str {
		dataBytes[i] = byte(c)
	}
	return dataBytes[:]
}

// 序列化对象
func Serialize(data interface{}) []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	if err := encoder.Encode(data); err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}
