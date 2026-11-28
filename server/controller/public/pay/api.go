package pay

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"server/global"
	"server/pkg/bean"
	"server/repository"
	paycommon "server/service/common/pay"
	payservice "server/service/normal/pay"
)

func collectParams(c *gin.Context) map[string]string {
	params := map[string]string{}
	_ = c.Request.ParseForm()
	for k, v := range c.Request.PostForm {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}
	if len(params) == 0 {
		var jsonParams map[string]string
		if err := c.ShouldBindJSON(&jsonParams); err == nil {
			for k, v := range jsonParams {
				params[k] = v
			}
		}
	}
	return params
}

// Notify 支付渠道异步回调入口（公开，无需登录）。验签→结算→应答 success/fail。
func Notify(c *gin.Context) {
	provider, ok := paycommon.Get(c.Param("provider"))
	if !ok {
		c.String(http.StatusBadRequest, "unknown provider")
		return
	}
	res := provider.VerifyNotify(collectParams(c))
	if !res.OK {
		c.String(http.StatusOK, "fail")
		return
	}
	if err := payservice.SettleOrder(res.OutTradeNo, res.TradeNo, res.Amount); err != nil {
		c.String(http.StatusOK, "fail")
		return
	}
	c.String(http.StatusOK, "success")
}

func sandboxKey() string {
	if global.Config.Pay.SandboxKey != "" {
		return global.Config.Pay.SandboxKey
	}
	return paycommon.DefaultSandboxKey
}

// SandboxPage 沙箱收银台（模拟支付页），供联调手动确认。
func SandboxPage(c *gin.Context) {
	code := c.Query("order")
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`<!doctype html><html><head><meta charset="utf-8"><title>沙箱支付</title></head>
<body style="font-family:sans-serif;max-width:480px;margin:60px auto;text-align:center">
<h2>沙箱收银台</h2><p>订单号：`+code+`</p>
<button id="pay" style="padding:10px 28px;font-size:16px">模拟支付成功</button>
<p id="msg"></p>
<script>
document.getElementById('pay').onclick=async()=>{
  const r=await fetch('/api/v1/public/pay/sandbox/confirm',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({code:'`+code+`'})});
  const j=await r.json();document.getElementById('msg').innerText=j.code===0?'支付成功，积分已到账':'失败：'+(j.msg||'');
};
</script></body></html>`))
}

// SandboxConfirm 沙箱模拟回调：构造带签名的回调参数走与真实渠道一致的验签+结算路径。
func SandboxConfirm(c *gin.Context) {
	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bean.Response.Param(c, err)
		return
	}
	db, _, _ := repository.Get("")
	order, err := db.SystemOrder.Where(db.SystemOrder.Code.Eq(req.Code)).First()
	if err != nil {
		bean.Response.Fail(c, "订单不存在")
		return
	}
	provider, ok := paycommon.Get(order.Provider)
	if !ok {
		bean.Response.Fail(c, "渠道不存在")
		return
	}
	amount := order.Amount.String()
	params := map[string]string{
		"out_trade_no": order.Code,
		"amount":       amount,
		"trade_no":     "sandbox_" + order.Code,
		"sign":         paycommon.SignSandbox(sandboxKey(), order.Code, amount),
	}
	res := provider.VerifyNotify(params)
	if !res.OK {
		bean.Response.Fail(c, "验签失败")
		return
	}
	if err := payservice.SettleOrder(res.OutTradeNo, res.TradeNo, res.Amount); err != nil {
		bean.Response.Fail(c, err.Error())
		return
	}
	bean.Response.Ok(c)
}
