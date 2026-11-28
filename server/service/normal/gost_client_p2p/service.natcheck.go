package service

import (
	"errors"

	"server/pkg/jwt"
	"server/repository"
)

// 默认 STUN 服务器（访客客户端用于 NAT 类型探测）。
var defaultStunServers = []string{
	"stun.qq.com:3478",
	"stun.miwifi.com:3478",
	"stun.chat.bilibili.com:3478",
}

type NatCheckReq struct {
	Code    string `json:"code" binding:"required"`
	NatType string `json:"natType"` // 客户端探测后上报：EasyNAT / SymmetricNAT / 空(未知)
}

type NatCheckResp struct {
	StunServers   []string `json:"stunServers"`
	NatType       string   `json:"natType"`
	DirectAdvice  string   `json:"directAdvice"`
	RelayFallback bool     `json:"relayFallback"`
	RelayBytes    int64    `json:"relayBytes"`
}

// NatCheck 返回 NAT 预检所需 STUN 服务器与直连建议；客户端上报 NatType 时记录。
// STUN 不可达时 NatType 为空，标记未知但不阻塞隧道使用（中继保底）。
func (service *service) NatCheck(claims jwt.Claims, req NatCheckReq) (NatCheckResp, error) {
	db, _, _ := repository.Get("")
	p2p, err := db.GostClientP2P.Where(
		db.GostClientP2P.Code.Eq(req.Code),
		db.GostClientP2P.UserCode.Eq(claims.Code),
	).First()
	if err != nil {
		return NatCheckResp{}, errors.New("隧道不存在")
	}

	natType := p2p.NatType
	if req.NatType != "" {
		switch req.NatType {
		case "EasyNAT", "SymmetricNAT", "Unknown":
			natType = req.NatType
			p2p.NatType = natType
			_ = db.GostClientP2P.Save(p2p)
		default:
			return NatCheckResp{}, errors.New("NAT类型无效")
		}
	}

	resp := NatCheckResp{
		StunServers:   defaultStunServers,
		NatType:       natType,
		RelayFallback: true,
		RelayBytes:    p2p.RelayBytes,
	}
	switch natType {
	case "EasyNAT":
		resp.DirectAdvice = "锥形NAT，直连打洞成功率高，优先尝试XTCP直连。"
	case "SymmetricNAT":
		resp.DirectAdvice = "对称NAT，直连打洞困难，建议启用STCP中继保证连通。"
		resp.RelayFallback = true
	default:
		resp.DirectAdvice = "NAT类型未知（STUN不可达或未探测），不阻塞使用，中继保底连通。"
	}
	return resp, nil
}
