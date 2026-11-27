package utils

import (
	"encoding/binary"
	"strconv"

	"github.com/gogf/gf/v2/util/gconv"
	"github.com/shopspring/decimal"
)

// 本文件：浮点运算与类型转换（对齐 EFunc 的 Float64* / Int64到Float64 / Int32ToBytes / D到* 系列）
//
// 浮点族内部用 shopspring/decimal 运算，避免 0.1+0.2=0.30000000000000004 这类
// 金额误差——主项目 go.mod 里本来就依赖 decimal，EFunc 的实现方式应当一致。

// 以下浮点族签名与舍入行为均由原版二进制反汇编逐条还原（EFunc/utils/Float64转换.go）：
//   Round 用的是 decimal.Decimal.Round，语义为 half away from zero（四舍五入，且远离零）
//   乘法两个变体都不做任何舍入，直接 Float64() 回落
//   除法走 DivRound(decimal.DivisionPrecision)（默认 16 位中间精度）后再 Round

// Float64到文本 按固定小数位输出（原版 Float64转换.go:75 等价于 strconv.FormatFloat(值,'f',小数位,64)）
func Float64到文本(值 float64, 小数位 int) string {
	return strconv.FormatFloat(值, 'f', 小数位, 64)
}

// Float64加float64 加法，结果按指定小数位四舍五入（原版 Float64转换.go:105-110）。
func Float64加float64(左 float64, 右 float64, 小数位 int) float64 {
	值, _ := decimal.NewFromFloat(左).Add(decimal.NewFromFloat(右)).Round(int32(小数位)).Float64()
	return 值
}

// Float64减float64 减法，结果按指定小数位四舍五入（原版 Float64转换.go:96-101）。
func Float64减float64(左 float64, 右 float64, 小数位 int) float64 {
	值, _ := decimal.NewFromFloat(左).Sub(decimal.NewFromFloat(右)).Round(int32(小数位)).Float64()
	return 值
}

// Float64乘Float64 乘法，只有两个参数且不做舍入（原版 Float64转换.go:32-38，全项目仅 agent.go:50 一处调用）。
func Float64乘Float64(左 float64, 右 float64) float64 {
	值, _ := decimal.NewFromFloat(左).Mul(decimal.NewFromFloat(右)).Float64()
	return 值
}

// Float64乘int64 浮点乘整数，不做舍入（原版 Float64转换.go:22-28）。
func Float64乘int64(左 float64, 右 int64) float64 {
	值, _ := decimal.NewFromFloat(左).Mul(decimal.NewFromInt(右)).Float64()
	return 值
}

// Float64除int64 浮点除整数，先 DivRound 再按小数位四舍五入（原版 Float64转换.go:42-48）。
// decimal.Div 内部即 DivRound(d2, DivisionPrecision)，与原版等价。
// 注：原版无除零保护，除以 0 会 panic；此处保留返回 0 的保护，避免线上请求直接崩溃。
func Float64除int64(左 float64, 右 int64, 小数位 int) float64 {
	if 右 == 0 {
		return 0
	}
	值, _ := decimal.NewFromFloat(左).Div(decimal.NewFromInt(右)).Round(int32(小数位)).Float64()
	return 值
}

// Float64取负值 只对正数取负，0 与负数原样返回（原版 Float64转换.go:62-71 的真实行为）。
func Float64取负值(值 float64) float64 {
	if 值 > 0 {
		结果, _ := decimal.NewFromFloat(值).Mul(decimal.NewFromInt(-1)).Float64()
		return 结果
	}
	return 值
}

// Float64取绝对值 只对负数取绝对，0 与正数原样返回（原版 Float64转换.go:9-18）。
func Float64取绝对值(值 float64) float64 {
	if 值 < 0 {
		结果, _ := decimal.NewFromFloat(值).Mul(decimal.NewFromInt(-1)).Float64()
		return 结果
	}
	return 值
}

// Int64到Float64 整数转浮点。
func Int64到Float64(值 int64) float64 {
	return float64(值)
}

// Int32ToBytes 把 int32 按小端序转为 4 字节（用于授权码/产品代码计算）。
func Int32ToBytes(值 int32) []byte {
	结果 := make([]byte, 4)
	binary.LittleEndian.PutUint32(结果, uint32(值))
	return 结果
}

// D到整数 任意值转整数。原版核心库.go:22 直接包装 gconv.Int64，此处保持同一实现，
// 以确保 "12.9"→12、bool→1、nil→0 等宽松转换细节与原版逐位一致。
func D到整数(值 interface{}) int {
	return gconv.Int(值)
}

// D到文本 任意值转文本。原版核心库.go:33 包装 gconv.String。
func D到文本(值 interface{}) string {
	return gconv.String(值)
}

// D到数值 任意值转浮点。原版核心库.go:30 包装 gconv.Float64。
func D到数值(值 interface{}) float64 {
	return gconv.Float64(值)
}
