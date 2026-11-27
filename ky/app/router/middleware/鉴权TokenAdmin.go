package middleware

import (
	"github.com/gin-gonic/gin"
	"server/app/global"
	"server/app/logic/common/lastTimeBuffer"
	"server/app/models/constant"
	"server/app/models/dbm"
	"server/app/models/old/response"
	"time"
)

// Token有效的才放行,否则返回Ttoken失效
func IsTokenAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		Token := c.Request.Header.Get("Token")

		if Token == "" {
			response.FailTokenErr(gin.H{"reload": true}, "请先登录", c)
			c.Abort()
			return
		}

		if global.GVA_DB == nil {
			response.FailTokenErr(gin.H{"reload": true}, "数据库连接失败,请重新设置", c)
			c.Abort()
			return
		}

		var DB_LinksToken dbm.DB_LinksToken
		//这里如果报错  invalid memory address or nil pointer dereference   可能是配置文件数据库配置北山,global.GVA_DB 值为空
		db := *global.GVA_DB
		err := db.Model(dbm.DB_LinksToken{}).Where("Token = ?", Token).First(&DB_LinksToken).Error
		// 没查到数据 或状态不正常
		if err != nil || DB_LinksToken.Status != 1 {
			response.FailTokenErr(gin.H{"reload": true}, "令牌已失效", c)
			c.Abort()
			return
		}

		if DB_LinksToken.LoginAppid != constant.APPID_管理平台 {
			response.FailTokenErr(gin.H{"reload": true}, "非管理员后台令牌,请重新登录.", c)
			c.Abort()
			return
		}
		//fmt.Println(DB_LinksToken)
		//更新最后活动时间
		if time.Now().Unix()-DB_LinksToken.LastTime > 60 { //超过1分钟,更新最后活动时间
			// LastTime 走本地缓冲,由定时任务批量回写DB,避免每请求UPDATE热表
			lastTimeBuffer.X心跳_记录(DB_LinksToken.Id)
		}
		//把 userID 保存到上下文,这样逻辑层就不用再查询了
		c.Set("Uid", DB_LinksToken.Uid)
		c.Set("User", DB_LinksToken.User)
		c.Set("局_在线信息", DB_LinksToken)

		if global.GVA_Viper.GetInt("系统模式") == 1 { //演示模式不强制跳
			c.Next()
			return
		}

		// 原版在此强制要求快验作者账号已登录：H会员帐号 为空时除 kuaiYan 相关接口外
		// 一律返回 KuaiYan:true 并把前端顶回「个人中心」登录框。
		// 自建部署没有快验账号体系，该页面已移除，故此校验一并去掉，
		// 后台仅依赖管理员自己的登录令牌(IsTokenAdmin 上方的 DB_LinksToken 校验)。

		/*if strings.Index(c.Request.RequestURI, "KuaiYan") == -1 && strings.Index(c.Request.RequestURI, "GetAdminInfo") == -1 && global.X系统信息.D到期时间 < time.Now().Unix() {
			局_计数 := 0
			局_计数缓存, ok := global.H缓存.Get("在线数量")
			//在个人中心那里获取就可以了,如果超过100 写入缓存,这样不影响速度
			if !ok {
				//如果没有写入缓存,就永远是0,直接放行
				局_计数 = 0
			} else {
				局_计数 = 局_计数缓存.(int)
			}
			//"恭喜管理员同时在线用户数量已超过100,请开通商业会员,感谢您的支持"
			局_临时, _ := base64.StdEncoding.DecodeString("5oGt5Zac566h55CG5ZGY5ZCM5pe25Zyo57q/55So5oi35pWw6YeP5bey6LaF6L+HMTAwLOivt+W8gOmAmuWVhuS4muS8muWRmCzmhJ/osKLmgqjnmoTmlK/mjIE=")
			if 局_计数 > 120 { //做一个容错处理 容错20
				response.FailTokenErr(gin.H{"KuaiYan": true}, string(局_临时), c)
				c.Abort()
				return
			}

		}*/

		// 继续处理请求
		c.Next()
	}
}
