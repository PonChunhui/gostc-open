package pay

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestSignSandboxDeterministic(t *testing.T) {
	a := SignSandbox("k", "order1", "10.50")
	b := SignSandbox("k", "order1", "10.50")
	if a != b || a == "" {
		t.Fatalf("sign not deterministic: %q vs %q", a, b)
	}
	if SignSandbox("k2", "order1", "10.50") == a {
		t.Fatal("different key should produce different sign")
	}
}

func TestSandboxVerifyNotify(t *testing.T) {
	s := &Sandbox{Key: "test-key"}
	amount := "12.34"
	good := map[string]string{
		"out_trade_no": "ord-1",
		"amount":       amount,
		"trade_no":     "wx-999",
		"sign":         SignSandbox("test-key", "ord-1", amount),
	}
	cases := []struct {
		name   string
		params map[string]string
		wantOK bool
	}{
		{"合法签名", good, true},
		{"错误签名", map[string]string{"out_trade_no": "ord-1", "amount": amount, "sign": "bad"}, false},
		{"缺签名", map[string]string{"out_trade_no": "ord-1", "amount": amount}, false},
		{"缺订单号", map[string]string{"amount": amount, "sign": SignSandbox("test-key", "", amount)}, false},
		{"金额被篡改", map[string]string{"out_trade_no": "ord-1", "amount": "99.99", "sign": good["sign"]}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res := s.VerifyNotify(c.params)
			if res.OK != c.wantOK {
				t.Fatalf("OK=%v want %v", res.OK, c.wantOK)
			}
			if c.wantOK {
				if res.OutTradeNo != "ord-1" || res.TradeNo != "wx-999" {
					t.Fatalf("unexpected result %+v", res)
				}
				if !res.Amount.Equal(decimal.RequireFromString(amount)) {
					t.Fatalf("amount=%s want %s", res.Amount, amount)
				}
			}
		})
	}
}
