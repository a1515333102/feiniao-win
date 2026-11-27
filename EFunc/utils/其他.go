package utils

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	qrcode "github.com/skip2/go-qrcode"
)

// 本文件：编码 / 时间 / 校验 / 二维码 / 随机数
// 对齐 EFunc 的 B编码_*、Z字节集_*、S时间_*、J校验_*、T图片_*、H汇编_* 系列

// B编码_BASE64编码 标准 Base64 编码。
func B编码_BASE64编码(内容 []byte) string {
	return base64.StdEncoding.EncodeToString(内容)
}

// B编码_BASE64解码 依次尝试标准/URL 编码、有无填充四种变体，全部失败返回 nil。
func B编码_BASE64解码(内容 string) []byte {
	内容 = strings.TrimSpace(内容)
	if 内容 == "" {
		return nil
	}
	编码表 := []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	}
	for _, 编码 := range 编码表 {
		if 结果, err := 编码.DecodeString(内容); err == nil {
			return 结果
		}
	}
	return nil
}

// Z字节集_字节集到十六进制 字节集转小写十六进制文本。
func Z字节集_字节集到十六进制(内容 []byte) string {
	return hex.EncodeToString(内容)
}

// Z字节集_十六进制到字节集 十六进制文本转字节集。
func Z字节集_十六进制到字节集(内容 string) []byte {
	内容 = strings.TrimSpace(内容)
	if 内容 == "" {
		return nil
	}
	结果, err := hex.DecodeString(内容)
	if err != nil {
		return nil
	}
	return 结果
}

// S时间_取现行时间戳 当前 Unix 时间戳（秒）。
func S时间_取现行时间戳() int64 {
	return time.Now().Unix()
}

// S时间_取现行时间戳13 当前 Unix 时间戳（毫秒，13 位）。
func S时间_取现行时间戳13() int64 {
	return time.Now().UnixMilli()
}

// 时间文本格式
const 时间格式 = "2006-01-02 15:04:05"

// S时间_时间戳到时间 时间戳格式化为 "2006-01-02 15:04:05"。
// 自动识别 10 位（秒）与 13 位（毫秒）；时间戳为 0 时返回空串。
func S时间_时间戳到时间(时间戳 int64) string {
	return S时间_时间戳到时间格式(时间戳, 时间格式)
}

// S时间_时间戳到时间格式 指定格式输出的版本。
func S时间_时间戳到时间格式(时间戳 int64, 格式 string) string {
	if 时间戳 <= 0 {
		return ""
	}
	var 时刻 time.Time
	if 时间戳 > 1e12 {
		时刻 = time.UnixMilli(时间戳)
	} else {
		时刻 = time.Unix(时间戳, 0)
	}
	if 格式 == "" {
		格式 = 时间格式
	}
	return 时刻.Format(格式)
}

// S时间_秒转时间文本 把秒数转成"X天X小时"这样的可读文本。
func S时间_秒转时间文本(秒 int64) string {
	if 秒 <= 0 {
		return "0秒"
	}
	常量 := []struct {
		单位 string
		秒数 int64
	}{
		{"天", 86400},
		{"小时", 3600},
		{"分", 60},
		{"秒", 1},
	}
	剩余 := 秒
	片段 := make([]string, 0, 2)
	已取 := 0
	for _, 项 := range 常量 {
		if 剩余 < 项.秒数 {
			continue
		}
		数值 := 剩余 / 项.秒数
		剩余 -= 数值 * 项.秒数
		片段 = append(片段, fmt.Sprintf("%d%s", 数值, 项.单位))
		已取++
		if 已取 >= 2 {
			break
		}
	}
	if len(片段) == 0 {
		return "0秒"
	}
	return strings.Join(片段, "")
}

// J校验_取md5_文本 取文本的 MD5（32 位小写十六进制）。
// 第二个参数为是否输出大写。
func J校验_取md5_文本(文本 string, 大写 ...bool) string {
	摘要 := md5.Sum([]byte(文本))
	结果 := hex.EncodeToString(摘要[:])
	if len(大写) > 0 && 大写[0] {
		return strings.ToUpper(结果)
	}
	return 结果
}

// T图片_生成二维码base64 把文本生成二维码 PNG，返回 data URI（可直接用于 img src）。
func T图片_生成二维码base64(内容 string) string {
	png字节, err := qrcode.Encode(内容, qrcode.Medium, 256)
	if err != nil {
		return ""
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(png字节)
}

// H汇编_取随机数 取 [最小, 最大] 闭区间内的随机整数。
func H汇编_取随机数(最小 int, 最大 int) int {
	if 最大 < 最小 {
		最小, 最大 = 最大, 最小
	}
	if 最大 == 最小 {
		return 最小
	}
	return 最小 + rand.IntN(最大-最小+1)
}
