// Package tunnel_adv 提供 frp 高级能力（负载均衡/健康检查/复用器/范围端口/带宽）表单字段的前置校验。
package tunnel_adv

import (
	"errors"
	"strconv"
	"strings"
)

// ValidateHealthCheck 校验健康检查参数：类型为空表示不启用；启用时 tcp/http，间隔必须>0。
func ValidateHealthCheck(hcType, path string, interval, timeout, maxFailed int) error {
	if hcType == "" {
		return nil
	}
	if hcType != "tcp" && hcType != "http" {
		return errors.New("健康检查类型仅支持 tcp 或 http")
	}
	if interval <= 0 {
		return errors.New("健康检查间隔必须大于 0 秒")
	}
	if timeout < 0 {
		return errors.New("健康检查超时不能为负")
	}
	if maxFailed < 0 {
		return errors.New("健康检查最大失败次数不能为负")
	}
	if hcType == "http" && strings.TrimSpace(path) == "" {
		return errors.New("HTTP 健康检查必须填写路径")
	}
	return nil
}

// ValidatePortRange 校验范围端口 "8000-8010"；空表示不启用；最多 1000 个端口。
func ValidatePortRange(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.SplitN(s, "-", 2)
	if len(parts) != 2 {
		return errors.New("范围端口格式应为 起始-结束，如 8000-8010")
	}
	start, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	end, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err1 != nil || err2 != nil || start <= 0 || end < start {
		return errors.New("范围端口非法")
	}
	if end-start > 1000 {
		return errors.New("范围端口数量不能超过 1000")
	}
	return nil
}

// ValidateMultiplexer 校验复用器类型：空表示普通 TCP；否则 httpconnect/tcpmux。
func ValidateMultiplexer(m string) error {
	if m == "" || m == "httpconnect" || m == "tcpmux" {
		return nil
	}
	return errors.New("复用器仅支持 httpconnect 或 tcpmux")
}

// ValidateBandwidth 校验带宽限制（如 1MB/512KB）与限制端（client/server）。
func ValidateBandwidth(limit, mode string) error {
	if mode != "" && mode != "client" && mode != "server" {
		return errors.New("带宽限制端仅支持 client 或 server")
	}
	limit = strings.TrimSpace(limit)
	if limit == "" {
		return nil
	}
	// 形如 <数字><单位>，单位 KB/MB/GB
	upper := strings.ToUpper(limit)
	unit := ""
	for _, u := range []string{"KB", "MB", "GB"} {
		if strings.HasSuffix(upper, u) {
			unit = u
			break
		}
	}
	if unit == "" {
		return errors.New("带宽限制格式应为 数字+单位，如 1MB、512KB")
	}
	numStr := strings.TrimSpace(upper[:len(upper)-len(unit)])
	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil || num <= 0 {
		return errors.New("带宽限制数值必须大于 0")
	}
	return nil
}
