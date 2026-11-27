package utils

import "regexp"

// 本文件：正则校验（对齐 EFunc 的 Z正则_* 系列）
//
// 签名沿用 EFunc 的 (值 string, 错误提示 *string) bool 形式：
// 校验失败时把提示写入 *msg，返回 false。
// 实现与主项目 app/utils/正则.go 保持一致——那份是作者手工从 EFunc 抄出来的，
// 正则表达式直接复用可确保行为相同。

// Z正则_校验密码 长度 5-18，只能包含非空白字符。
func Z正则_校验密码(值 string, 提示 *string) bool {
	匹配结果, _ := regexp.MatchString(`\S{5,18}$`, 值)
	if !匹配结果 && 提示 != nil {
		*提示 = "长度在5-18之间，只能包含字母、数字和下划线等非空白字符"
	}
	return 匹配结果
}

// Z正则_校验代理用户名 只能包含英文字母、数字和中文字符。
func Z正则_校验代理用户名(值 string, 提示 *string) bool {
	匹配结果, _ := regexp.MatchString(`^[a-zA-Z0-9\p{Han}]+$`, 值)
	if !匹配结果 && 提示 != nil {
		*提示 = "只能包含英文字母、数字和ANSI编码支持的中文字符"
	}
	return 匹配结果
}

// Z正则_校验用户名 长度 5-17，只能包含字符、数字和下划线。
func Z正则_校验用户名(值 string, 提示 *string) bool {
	匹配结果, _ := regexp.MatchString(`\w{5,17}$`, 值)
	if 匹配结果 {
		return 匹配结果
	}
	if 提示 != nil {
		*提示 = "长度在6-18之间，只能包含字符、数字和下划线"
	}
	return 匹配结果
}

// Z正则_校验email 邮箱格式校验。
func Z正则_校验email(值 string, 提示 *string) bool {
	匹配结果, _ := regexp.MatchString(`^\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*$`, 值)
	if 匹配结果 {
		return 匹配结果
	}
	if 提示 != nil {
		*提示 = "非正确email格式"
	}
	return 匹配结果
}
