package utils

import (
	"encoding/json"
	"math/rand/v2"
	"regexp"
	"strings"
	"unicode/utf8"
)

// 本文件：文本处理（对齐 EFunc 的 W文本_* 系列）
//
// 注意易语言的语义差异，这几个函数容易混：
//   W文本_取左边(s, 4)      -> 按"字符数"取左边 4 个字符
//   W文本_取文本左边(s, "/") -> 按"分隔符"取左边部分
//   取右边 同理。

const (
	随机字符_大小写数字 = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	随机字符_数字      = "0123456789"
)

// W文本_是否包含关键字 源文本是否包含任意一个关键字（任一命中即 true）。
func W文本_是否包含关键字(源文本 string, 关键字 ...string) bool {
	for _, k := range 关键字 {
		if k == "" {
			continue
		}
		if strings.Contains(源文本, k) {
			return true
		}
	}
	return false
}

// W文本_分割文本 按分隔符切分为文本数组；分隔符为空时按字符切分。
func W文本_分割文本(源文本 string, 分隔符 string) []string {
	if 分隔符 == "" {
		return strings.Split(源文本, "")
	}
	return strings.Split(源文本, 分隔符)
}

// W文本_取随机字符串 取指定长度的随机字符串（大小写字母 + 数字）。
func W文本_取随机字符串(长度 int) string {
	return 取随机字符(长度, 随机字符_大小写数字)
}

// W文本_取随机字符串_数字 取指定长度的随机数字串。
func W文本_取随机字符串_数字(长度 int) string {
	return 取随机字符(长度, 随机字符_数字)
}

func 取随机字符(长度 int, 字符集 string) string {
	if 长度 <= 0 || 字符集 == "" {
		return ""
	}
	b := make([]byte, 长度)
	for i := range b {
		b[i] = 字符集[rand.IntN(len(字符集))]
	}
	return string(b)
}

// W文本_取出中间文本 取出 左文本 与 紧邻其后的 右文本 之间的内容。
// 左文本为空则从头取；右文本为空则取到结尾；找不到返回空串。
func W文本_取出中间文本(源文本 string, 左文本 string, 右文本 string) string {
	起始 := 0
	if 左文本 != "" {
		位置 := strings.Index(源文本, 左文本)
		if 位置 < 0 {
			return ""
		}
		起始 = 位置 + len(左文本)
	}
	if 起始 > len(源文本) {
		return ""
	}
	剩余 := 源文本[起始:]
	if 右文本 == "" {
		return 剩余
	}
	结束 := strings.Index(剩余, 右文本)
	if 结束 < 0 {
		return ""
	}
	return 剩余[:结束]
}

// W文本_去除敏感信息 日志脱敏。
// 算法由原版二进制 W文本.go:592-617 反汇编还原，与"保留首尾固定位数"的常见写法不同：
//   n<=2：n==1 返回 "*"；n==2 返回 首字符+"*"（n==0 原版会数组越界 panic，此处返回空串）
//   n>2 ：半长=n/2，起点=(n-半长)/2，把 [起点, 起点+半长) 区间整段替换为 '*'，首尾长度随 n 变化
// 例：n=10 → "ab*****hij"（前2后3），n=11 → "abc*****hij"（前3后3）。
func W文本_去除敏感信息(源文本 string) string {
	字符 := []rune(源文本)
	长 := len(字符)
	if 长 <= 2 {
		switch 长 {
		case 0:
			return ""
		case 1:
			return "*"
		default:
			return string(字符[0]) + "*"
		}
	}
	半 := 长 / 2
	起 := (长 - 半) / 2
	for i := 起; i < 起+半; i++ {
		字符[i] = '*'
	}
	return string(字符)
}

// W文本_取左边 按字符数取左边 N 个字符（易语言"取文本左边"语义）。
func W文本_取左边(源文本 string, 长度 int) string {
	if 长度 <= 0 {
		return ""
	}
	字符 := []rune(源文本)
	if 长度 >= len(字符) {
		return 源文本
	}
	return string(字符[:长度])
}

// W文本_取右边 按字符数取右边 N 个字符。
func W文本_取右边(源文本 string, 长度 int) string {
	if 长度 <= 0 {
		return ""
	}
	字符 := []rune(源文本)
	if 长度 >= len(字符) {
		return 源文本
	}
	return string(字符[len(字符)-长度:])
}

// W文本_取文本左边 取第一个分隔符之前的部分；无分隔符时返回原文本。
func W文本_取文本左边(源文本 string, 分隔符 string) string {
	if 分隔符 == "" {
		return 源文本
	}
	位置 := strings.Index(源文本, 分隔符)
	if 位置 < 0 {
		return 源文本
	}
	return 源文本[:位置]
}

// W文本_取文本右边 取最后一个分隔符之后的部分；无分隔符时返回原文本。
func W文本_取文本右边(源文本 string, 分隔符 string) string {
	if 分隔符 == "" {
		return 源文本
	}
	位置 := strings.LastIndex(源文本, 分隔符)
	if 位置 < 0 {
		return 源文本
	}
	return 源文本[位置+len(分隔符):]
}

// W文本_删首尾空 去掉首尾空白字符。
func W文本_删首尾空(源文本 string) string {
	return strings.TrimSpace(源文本)
}

// W文本_取长度 取字符数（按 Unicode 字符计，与易语言一致，不是字节数）。
func W文本_取长度(源文本 string) int {
	return utf8.RuneCountInString(源文本)
}

// W文本_是否为数字 是否全部由 0-9 组成（空串返回 false）。
func W文本_是否为数字(源文本 string) bool {
	if 源文本 == "" {
		return false
	}
	for _, r := range 源文本 {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// W文本_可能为json 是否是合法的 JSON 对象或数组文本。
func W文本_可能为json(源文本 string) bool {
	裁剪 := strings.TrimSpace(源文本)
	if 裁剪 == "" {
		return false
	}
	if 裁剪[0] != '{' && 裁剪[0] != '[' {
		return false
	}
	var 临时 interface{}
	return json.Unmarshal([]byte(裁剪), &临时) == nil
}

// W文本_是否存在_任意 源文本是否包含关键字数组中的任意一个。
func W文本_是否存在_任意(源文本 string, 关键字数组 []string) bool {
	return W文本_是否包含关键字(源文本, 关键字数组...)
}

// W文本_取行数 文本行数；空文本为 0。
func W文本_取行数(源文本 string) int {
	if 源文本 == "" {
		return 0
	}
	return strings.Count(源文本, "\n") + 1
}

// Z正则_取全部匹配子文本 按正则取出全部匹配到的子文本。
func Z正则_取全部匹配子文本(源文本 string, 表达式 string) []string {
	编译结果, err := regexp.Compile(表达式)
	if err != nil {
		return []string{}
	}
	匹配 := 编译结果.FindAllString(源文本, -1)
	if 匹配 == nil {
		return []string{}
	}
	return 匹配
}

// W文本_子文本替换 替换文本。
// 可选参数：[起始位置, 替换次数]，替换次数为 -1 或省略时全部替换。
func W文本_子文本替换(源文本 string, 被替换文本 string, 替换文本 string, 参数 ...int) string {
	if 被替换文本 == "" {
		return 源文本
	}
	if len(参数) >= 2 && 参数[1] >= 0 {
		return strings.Replace(源文本, 被替换文本, 替换文本, 参数[1])
	}
	return strings.ReplaceAll(源文本, 被替换文本, 替换文本)
}
