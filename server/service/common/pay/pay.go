// Package pay 提供可插拔支付网关抽象。本期仅实现沙箱渠道（框架就绪），
// 真实渠道（支付宝/微信/Stripe）保留接口与配置位，未配置商户资质时拒绝下单。
package pay

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"

	"github.com/shopspring/decimal"
	"server/configs"
	"server/model"
)

// NotifyResult 异步回调验签结果。
type NotifyResult struct {
	OutTradeNo string
	TradeNo    string
	Amount     decimal.Decimal
	OK         bool
}

// Provider 统一支付网关接口。
type Provider interface {
	Name() string
	// CreateOrder 创建网关订单，返回支付跳转 URL。
	CreateOrder(order *model.SystemOrder) (string, error)
	// VerifyNotify 校验异步通知签名并解析订单号/交易号/金额。
	VerifyNotify(params map[string]string) NotifyResult
	// QueryOrder 主动查单（回调丢失补偿），返回订单状态与交易号。
	QueryOrder(outTradeNo string) (status int, tradeNo string, err error)
}

var (
	mu        sync.RWMutex
	providers = map[string]Provider{}
)

// Register 注册支付渠道。
func Register(p Provider) {
	mu.Lock()
	defer mu.Unlock()
	providers[p.Name()] = p
}

// Get 获取支付渠道。
func Get(name string) (Provider, bool) {
	mu.RLock()
	defer mu.RUnlock()
	p, ok := providers[name]
	return p, ok
}

// DefaultSandboxKey 沙箱默认签名密钥（配置未指定时使用）。
const DefaultSandboxKey = "gostc-sandbox"

// Init 按配置注册渠道：沙箱始终可用；真实渠道仅在配置商户资质后启用。
func Init(cfg configs.Pay) {
	key := cfg.SandboxKey
	if key == "" {
		key = DefaultSandboxKey
	}
	Register(&Sandbox{Key: key})
	// 真实渠道占位：未配置资质时 CreateOrder 返回错误，引导使用沙箱。
	Register(&stubProvider{name: "alipay", configured: cfg.Alipay.AppID != ""})
	Register(&stubProvider{name: "wechat", configured: cfg.Wechat.MchID != ""})
	Register(&stubProvider{name: "stripe", configured: cfg.Stripe.SecretKey != ""})
}

// SignSandbox 生成沙箱回调签名：HMAC-SHA256(outTradeNo|amount, key)。
func SignSandbox(key, outTradeNo, amount string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(outTradeNo + "|" + amount))
	return hex.EncodeToString(mac.Sum(nil))
}

// Sandbox 沙箱渠道：本地模拟下单与回调，用于打通全链路（无真实资金）。
type Sandbox struct {
	Key string
}

func (s *Sandbox) Name() string { return "sandbox" }

func (s *Sandbox) CreateOrder(order *model.SystemOrder) (string, error) {
	return fmt.Sprintf("/api/v1/public/pay/sandbox/page?order=%s", order.Code), nil
}

func (s *Sandbox) VerifyNotify(params map[string]string) NotifyResult {
	outTradeNo := params["out_trade_no"]
	amount := params["amount"]
	sign := params["sign"]
	if outTradeNo == "" || amount == "" || sign == "" {
		return NotifyResult{}
	}
	expect := SignSandbox(s.Key, outTradeNo, amount)
	if !hmac.Equal([]byte(expect), []byte(sign)) {
		return NotifyResult{}
	}
	amt, err := decimal.NewFromString(amount)
	if err != nil {
		return NotifyResult{}
	}
	tradeNo := params["trade_no"]
	if tradeNo == "" {
		tradeNo = "sandbox_" + outTradeNo
	}
	return NotifyResult{OutTradeNo: outTradeNo, TradeNo: tradeNo, Amount: amt, OK: true}
}

func (s *Sandbox) QueryOrder(outTradeNo string) (int, string, error) {
	// 沙箱查单默认未支付，结算以回调为准。
	return model.ORDER_STATUS_PENDING, "", nil
}

// stubProvider 真实渠道占位实现。
type stubProvider struct {
	name       string
	configured bool
}

func (p *stubProvider) Name() string { return p.name }

func (p *stubProvider) CreateOrder(order *model.SystemOrder) (string, error) {
	if !p.configured {
		return "", errors.New("该支付渠道未配置商户资质，请使用沙箱渠道")
	}
	return "", errors.New("真实支付渠道尚未接入，请使用沙箱渠道")
}

func (p *stubProvider) VerifyNotify(params map[string]string) NotifyResult {
	return NotifyResult{}
}

func (p *stubProvider) QueryOrder(outTradeNo string) (int, string, error) {
	return 0, "", errors.New("真实支付渠道尚未接入")
}
