package cache

import "testing"

func TestIsBlackIp(t *testing.T) {
	a := AdmissionInfo{BlackEnable: 1, BlackMap: map[string]struct{}{"1.2.3.4": {}}}
	if !a.IsBlackIp("1.2.3.4") {
		t.Fatal("listed ip should be black")
	}
	if a.IsBlackIp("5.6.7.8") {
		t.Fatal("unlisted ip should not be black")
	}
	disabled := AdmissionInfo{BlackEnable: 2, BlackMap: map[string]struct{}{"1.2.3.4": {}}}
	if disabled.IsBlackIp("1.2.3.4") {
		t.Fatal("disabled blacklist should not match")
	}
}

func TestValidWhiteIp(t *testing.T) {
	open := AdmissionInfo{WhiteEnable: 2}
	if !open.ValidWhiteIp("9.9.9.9") {
		t.Fatal("whitelist disabled should allow all")
	}
	a := AdmissionInfo{WhiteEnable: 1, WhiteMap: map[string]struct{}{"1.1.1.1": {}}}
	if !a.ValidWhiteIp("1.1.1.1") {
		t.Fatal("listed ip should be allowed")
	}
	if a.ValidWhiteIp("2.2.2.2") {
		t.Fatal("unlisted ip should be denied when whitelist enabled")
	}
}

func TestBlackOverridesWhite(t *testing.T) {
	// 同一 IP 同时在黑白名单：黑名单优先（enforcement 先判 IsBlackIp）。
	a := AdmissionInfo{
		WhiteEnable: 1, WhiteMap: map[string]struct{}{"1.1.1.1": {}},
		BlackEnable: 1, BlackMap: map[string]struct{}{"1.1.1.1": {}},
	}
	if !a.IsBlackIp("1.1.1.1") {
		t.Fatal("black should match even if also whitelisted")
	}
}
