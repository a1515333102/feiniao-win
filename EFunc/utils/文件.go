package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// 本文件：文件与目录 IO（对齐 EFunc 的 W文件_* / C程序_ / M目录_ 系列）

// W文件_是否存在 判断文件是否存在（是目录时返回 false）。
func W文件_是否存在(路径 string) bool {
	if 路径 == "" {
		return false
	}
	信息, err := os.Stat(路径)
	if err != nil {
		return false
	}
	return !信息.IsDir()
}

// W文件_读入文件 读取文件全部字节；失败返回 nil。
func W文件_读入文件(路径 string) []byte {
	内容, err := os.ReadFile(路径)
	if err != nil {
		return nil
	}
	return 内容
}

// W文件_读入文本 读取文件为文本；失败返回空串。
func W文件_读入文本(路径 string) string {
	return string(W文件_读入文件(路径))
}

// W文件_删除 删除文件；成功（或文件本就不存在）返回 nil。
func W文件_删除(路径 string) error {
	if 路径 == "" {
		return nil
	}
	err := os.Remove(路径)
	if err != nil && os.IsNotExist(err) {
		return nil
	}
	return err
}

// W文件_保存 以文本方式写入文件，自动创建上级目录。
func W文件_保存(路径 string, 内容 string) error {
	建父目录(路径)
	return os.WriteFile(路径, []byte(内容), 0o644)
}

// W文件_写出文件 以字节方式写入文件，自动创建上级目录。
func W文件_写出文件(路径 string, 内容 []byte) error {
	建父目录(路径)
	return os.WriteFile(路径, 内容, 0o644)
}

// W文件_写到文件 同 W文件_写出文件（源码中两个名字都在用）。
func W文件_写到文件(路径 string, 内容 []byte) error {
	return W文件_写出文件(路径, 内容)
}

func 建父目录(路径 string) {
	目录 := 取父目录(路径)
	if 目录 != "" {
		_ = os.MkdirAll(目录, 0o755)
	}
}

// W文件_取文件名 取路径最后一段（同时兼容 / 与 \ 分隔符）。
func W文件_取文件名(路径 string) string {
	清理 := strings.TrimRight(路径, `/\`)
	if 位置 := strings.LastIndexAny(清理, `/\`); 位置 >= 0 {
		return 清理[位置+1:]
	}
	return 清理
}

// W文件_取父目录 取上级目录路径。
func W文件_取父目录(路径 string) string {
	return 取父目录(路径)
}

func 取父目录(路径 string) string {
	清理 := strings.TrimRight(路径, `/\`)
	位置 := strings.LastIndexAny(清理, `/\`)
	if 位置 < 0 {
		return ""
	}
	if 位置 == 0 {
		return 清理[:1]
	}
	return 清理[:位置]
}

// C程序_取运行目录 取当前可执行文件所在目录（解析软链接后的真实路径）。
func C程序_取运行目录() string {
	执行文件, err := os.Executable()
	if err != nil {
		return "."
	}
	if 真实路径, err2 := filepath.EvalSymlinks(执行文件); err2 == nil {
		执行文件 = 真实路径
	}
	return filepath.Dir(执行文件)
}

// M目录_创建 递归创建目录，支持一次传多个。
func M目录_创建(目录 ...string) bool {
	成功 := true
	有传入 := false
	for _, d := range 目录 {
		if strings.TrimSpace(d) == "" {
			continue
		}
		有传入 = true
		if err := os.MkdirAll(d, 0o755); err != nil {
			成功 = false
		}
	}
	if !有传入 {
		return false
	}
	return 成功
}
