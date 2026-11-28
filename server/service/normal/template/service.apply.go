package service

import (
	"encoding/json"
	"errors"

	"server/model"
	"server/pkg/jwt"
	"server/repository"
	forwardservice "server/service/normal/gost_client_forward"
	hostservice "server/service/normal/gost_client_host"
)

type ApplyReq struct {
	Code       string `json:"code" binding:"required"`
	Name       string `json:"name"`
	NodeCode   string `json:"nodeCode"`
	ClientCode string `json:"clientCode"`
	ConfigCode string `json:"configCode"`
}

// Apply 用模板创建一条隧道：模板 Content 为对应类型的创建参数 JSON，可用入参覆盖名称/节点/客户端/套餐。
func (s *service) Apply(claims jwt.Claims, req ApplyReq) error {
	db, _, _ := repository.Get("")
	tpl, err := db.SystemTunnelTemplate.Where(db.SystemTunnelTemplate.Code.Eq(req.Code), db.SystemTunnelTemplate.UserCode.Eq(claims.Code)).First()
	if err != nil {
		return errors.New("模板不存在")
	}
	switch tpl.Kind {
	case model.TUNNEL_TEMPLATE_KIND_HOST:
		var create hostservice.CreateReq
		if err := json.Unmarshal([]byte(tpl.Content), &create); err != nil {
			return errors.New("模板内容解析失败")
		}
		if req.Name != "" {
			create.Name = req.Name
		}
		if req.NodeCode != "" {
			create.NodeCode = req.NodeCode
		}
		if req.ClientCode != "" {
			create.ClientCode = req.ClientCode
		}
		if req.ConfigCode != "" {
			create.ConfigCode = req.ConfigCode
		}
		return hostservice.Service.Create(claims, create)
	case model.TUNNEL_TEMPLATE_KIND_FORWARD:
		var create forwardservice.CreateReq
		if err := json.Unmarshal([]byte(tpl.Content), &create); err != nil {
			return errors.New("模板内容解析失败")
		}
		if req.Name != "" {
			create.Name = req.Name
		}
		if req.NodeCode != "" {
			create.NodeCode = req.NodeCode
		}
		if req.ClientCode != "" {
			create.ClientCode = req.ClientCode
		}
		if req.ConfigCode != "" {
			create.ConfigCode = req.ConfigCode
		}
		return forwardservice.Service.Create(claims, create)
	default:
		return errors.New("该类型模板暂不支持一键应用，请复制内容手动创建")
	}
}
