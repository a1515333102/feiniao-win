package rmbPay

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/imroc/req/v3"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"server/app/global"
	"server/app/logic/common/setting"
	"server/app/logic/webUser/cpsPayOrder"
	m "server/app/models/common"
	"server/app/models/constant"
	"server/app/models/dbm"
	utils2 "server/app/utils"
)

// 本文件实现「主动查单」：不依赖上游异步回调，定时向上游查询待支付订单的真实状态，
// 已支付的订单补记入账。
//
// 为什么需要它：
//   1. 上游异步通知要求 notify_url 公网可达，内网/本地调试时收不到；
//   2. 通知可能丢失、被防火墙拦截，或系统短时重启；
//   3. pay.10sn.cn 提供 GET /Api/checkOrder，可以直接查询。
//
// 入账流程与异步回调完全一致（行锁 + 状态判断 + Z支付成功_后处理），
// 因此即使回调与查单同时命中同一笔订单，也只会入账一次。

const (
	// C查单_通道名 易支付通道名称，与 rmbPayItem/易支付.go 的 Q取通道名称() 保持一致
	C查单_通道名 = "易支付"
	// C查单_最大单次 单次最多处理多少笔
	C查单_最大单次 = 50
	// C查单_延迟秒 订单创建后多少秒才开始查，给异步回调留一个先手窗口
	C查单_延迟秒 = 20
	// C查单_最长追查小时 超过这个时长仍未支付的订单不再追查
	C查单_最长追查小时 = 3
)

// K查单任务 定时向上游查询待支付订单状态，已支付的补记入账。
func (j *rmbPay) K查单任务() {
	if global.GVA_DB == nil {
		return
	}
	支付配置 := setting.Q在线支付配置()
	if !支付配置.Y易支付开关 ||
		支付配置.Y易支付网关 == "" ||
		支付配置.Y易支付商户ID == "" ||
		支付配置.Y易支付商户密钥KEY == "" {
		return
	}

	局_现在 := time.Now().Unix()
	tx := *global.GVA_DB
	var 订单列表 []dbm.DB_LogRMBPayOrder
	err := tx.Model(dbm.DB_LogRMBPayOrder{}).
		Where("Status = ? AND Type = ? AND Time < ? AND Time > ?",
			constant.D订单状态_等待支付,
			C查单_通道名,
			局_现在-C查单_延迟秒,
			局_现在-C查单_最长追查小时*3600).
		Order("Id asc").
		Limit(C查单_最大单次).
		Find(&订单列表).Error
	if err != nil {
		return
	}

	for i := range 订单列表 {
		局_订单 := 订单列表[i]
		已支付, 交易号, 实付金额, 局_错误 := j.Q查上游订单(支付配置, 局_订单.PayOrder)
		if 局_错误 != nil {
			global.GVA_LOG.Println("支付查单失败 " + 局_订单.PayOrder + ":" + 局_错误.Error())
			continue
		}
		if !已支付 {
			continue
		}
		j.Z补记已支付(&局_订单, 交易号, 实付金额)
	}
}

// Q查上游订单 调用上游 GET /Api/checkOrder 查询单笔订单。
// 返回：是否已支付、上游交易号、实付金额、错误。
func (j *rmbPay) Q查上游订单(配置 m.Z在线支付, 商户订单号 string) (已支付 bool, 交易号 string, 实付金额 float64, err error) {
	局_参数 := url.Values{}
	局_参数.Set("pid", 配置.Y易支付商户ID)
	局_参数.Set("out_trade_no", 商户订单号)
	局_参数.Set("sign", 易支付算签名(局_参数, 配置.Y易支付商户密钥KEY))

	局_请求 := req.C().SetTimeout(15 * time.Second).EnableInsecureSkipVerify().R()
	for k := range 局_参数 {
		局_请求.SetQueryParam(k, 局_参数.Get(k))
	}
	局_响应, 局_错误 := 局_请求.Get(易支付接口基址(配置.Y易支付网关) + "/Api/checkOrder")
	if 局_错误 != nil {
		return false, "", 0, 局_错误
	}
	if 局_响应 == nil || 局_响应.StatusCode != http.StatusOK {
		return false, "", 0, errors.New("查单接口返回异常")
	}

	局_json, 局_错误 := gjson.LoadJson(局_响应.String())
	if 局_错误 != nil {
		return false, "", 0, errors.New("查单结果解析失败")
	}
	// code=0 才是查询成功；订单不存在(未支付)等返回非 0，属于正常情况，不当成错误
	if 局_json.Get("code").Int() != 0 {
		return false, "", 0, nil
	}
	if 局_json.Get("data").IsNil() {
		return false, "", 0, nil
	}
	if 局_json.Get("data.status").Int() != 1 { // 1=已支付
		return false, "", 0, nil
	}
	return true, 局_json.Get("data.trade_no").String(), 局_json.Get("data.money").Float64(), nil
}

// Z补记已支付 把订单由「等待支付」置为「已付待处理」，并执行与异步回调相同的入账流程。
func (j *rmbPay) Z补记已支付(订单 *dbm.DB_LogRMBPayOrder, 交易号 string, 实付金额 float64) {
	var 参数 m.PayParams
	参数.Z支付配置s = setting.Q在线支付配置()
	参数.Z支付配置, _ = json.Marshal(&参数.Z支付配置s)
	参数.DB_LogRMBPayOrder = *订单

	局_额外信息, 局_错误 := gjson.LoadJson(参数.Extra)
	if 局_错误 != nil || 局_额外信息 == nil {
		局_额外信息, _ = gjson.LoadJson("{}")
	}
	参数.E额外信息 = 局_额外信息
	_ = 参数.E额外信息.Set("查单确认时间", time.Now().Unix())

	if 交易号 != "" {
		参数.PayOrder2 = 交易号
	}
	if 实付金额 > 0 {
		参数.ActualRmb = 实付金额
	}

	db := *global.GVA_DB
	// 与回调一致：先加行锁再判断状态，保证并发下只入账一次
	局_错误 = db.Transaction(func(tx *gorm.DB) error {
		var 局_订单信息 dbm.DB_LogRMBPayOrder
		if err := tx.Model(dbm.DB_LogRMBPayOrder{}).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("Id=?", 参数.Id).
			First(&局_订单信息).Error; err != nil {
			return err
		}
		if 局_订单信息.Status != constant.D订单状态_等待支付 {
			return errors.New("订单已被处理")
		}
		参数.Status = constant.D订单状态_已付待处理
		return tx.Model(dbm.DB_LogRMBPayOrder{}).
			Where("Id=?", 参数.Id).
			Updates(map[string]interface{}{
				"PayOrder2": 参数.PayOrder2,
				"Status":    参数.Status,
				"ActualRmb": 参数.ActualRmb,
			}).Error
	})
	if 局_错误 != nil {
		return
	}

	// 造一个最小可用的上下文：入账逻辑只用到 c.Set/c.Get，不依赖真实请求
	局_上下文, _ := gin.CreateTestContext(httptest.NewRecorder())
	局_上下文.Request, _ = http.NewRequest(http.MethodPost, "/", nil)

	_ = j.Z支付成功_后处理(局_上下文, &参数)
	_ = cpsPayOrder.L_cpsPayOrder.C处理佣金发放_线程安全(局_上下文, &参数)

	global.GVA_LOG.Println("查单确认支付成功:" + 参数.PayOrder)
}

// 易支付算签名 按参数名 ASCII 升序拼 k=v&k=v 后接商户 KEY 取 MD5，与易支付.go 一致。
func 易支付算签名(参数 url.Values, KEY string) string {
	局_键列表 := make([]string, 0, len(参数))
	for k := range 参数 {
		if k == "sign" || k == "sign_type" || 参数.Get(k) == "" {
			continue
		}
		局_键列表 = append(局_键列表, k)
	}
	sort.Strings(局_键列表)

	var 局_拼接 strings.Builder
	for i, k := range 局_键列表 {
		if i > 0 {
			局_拼接.WriteString("&")
		}
		局_拼接.WriteString(k)
		局_拼接.WriteString("=")
		局_拼接.WriteString(参数.Get(k))
	}
	return utils2.Md5String(局_拼接.String() + KEY)
}

// 易支付接口基址 取网关的 scheme://host，兼容用户填了 /submit.php 这类路径的情况。
func 易支付接口基址(网关 string) string {
	局_文本 := strings.TrimSpace(网关)
	局_url, 局_错误 := url.Parse(局_文本)
	if 局_错误 != nil || 局_url.Host == "" {
		return strings.TrimRight(局_文本, "/")
	}
	return 局_url.Scheme + "://" + 局_url.Host
}
