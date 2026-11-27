package utils

import "strings"

// 本文件：泛型语法糖 / 集合(Map、切片)工具
// 命名沿用 EFunc 的易语言风格，供 KuaiYanGo2 以点导入方式调用。

// S三元 三元表达式：条件成立取 真值，否则取 假值。
// 用泛型实现，任何类型都能用（源码中 46 处调用涉及 string/int64/其它）。
func S三元[T any](条件 bool, 真值 T, 假值 T) T {
	if 条件 {
		return 真值
	}
	return 假值
}

// Map_键名是否存在 判断 map 中是否存在指定键。
// 用泛型实现，兼容 map[string]string / map[string]interface{} / map[int]xxx 等各种形态。
func Map_键名是否存在[K comparable, V any](目标 map[K]V, 键 K) bool {
	_, ok := 目标[键]
	return ok
}

// S数组_整数是否存在 reports whether 值 exists in 数组.
func S数组_整数是否存在(数组 []int, 值 int) bool {
	for _, v := range 数组 {
		if v == 值 {
			return true
		}
	}
	return false
}

// S数组_去重复 去除重复元素，保持首次出现的顺序。
// 泛型实现，源码中同时用于 []string 与 []int。
func S数组_去重复[T comparable](数组 []T) []T {
	if len(数组) == 0 {
		return []T{}
	}
	已见 := make(map[T]struct{}, len(数组))
	结果 := make([]T, 0, len(数组))
	for _, v := range 数组 {
			if _, ok := 已见[v]; ok {
				continue
			}
			已见[v] = struct{}{}
		结果 = append(结果, v)
	}
	return 结果
}

// S数组_是否存在 判断元素是否存在于数组。
func S数组_是否存在[T comparable](数组 []T, 值 T) bool {
	for _, v := range 数组 {
		if v == 值 {
			return true
		}
	}
	return false
}

// S数组_取文本出现次数 统计指定文本在数组中出现的次数。
func S数组_取文本出现次数(数组 []string, 值 string) int {
	次数 := 0
	for _, v := range 数组 {
		if v == 值 {
			次数++
		}
	}
	return 次数
}

// S数组_取差集 返回存在于 数组A 但不存在于 数组B 的元素（A - B），保持 A 的顺序。
func S数组_取差集[T comparable](数组A []T, 数组B []T) []T {
	if len(数组A) == 0 {
		return []T{}
	}
	排除 := make(map[T]struct{}, len(数组B))
	for _, v := range 数组B {
		排除[v] = struct{}{}
	}
	结果 := make([]T, 0, len(数组A))
	for _, v := range 数组A {
		if _, ok := 排除[v]; !ok {
			结果 = append(结果, v)
		}
	}
	return 结果
}

// S数组_整数取差集 整数切片的差集（A - B）。
func S数组_整数取差集(数组A []int, 数组B []int) []int {
	return S数组_取差集(数组A, 数组B)
}

// S数组_合并文本 用分隔符把文本数组合并成一个字符串。
func S数组_合并文本(数组 []string, 分隔符 string) string {
	return strings.Join(数组, 分隔符)
}
