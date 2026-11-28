package service

import (
	"encoding/json"
	"errors"

	"server/model"
	"server/pkg/bean"
	"server/pkg/jwt"
	"server/repository"
)

type CreateReq struct {
	Name    string          `json:"name" binding:"required"`
	Kind    string          `json:"kind" binding:"required"`
	Content json.RawMessage `json:"content" binding:"required"`
}

func validKind(kind string) bool {
	switch kind {
	case model.TUNNEL_TEMPLATE_KIND_HOST, model.TUNNEL_TEMPLATE_KIND_FORWARD,
		model.TUNNEL_TEMPLATE_KIND_PROXY, model.TUNNEL_TEMPLATE_KIND_TUNNEL, model.TUNNEL_TEMPLATE_KIND_P2P:
		return true
	}
	return false
}

func (s *service) Create(claims jwt.Claims, req CreateReq) error {
	if !validKind(req.Kind) {
		return errors.New("隧道类型无效")
	}
	if !json.Valid(req.Content) {
		return errors.New("模板内容必须为合法JSON")
	}
	db, _, _ := repository.Get("")
	rec := &model.SystemTunnelTemplate{
		UserCode: claims.Code,
		Name:     req.Name,
		Kind:     req.Kind,
		Content:  string(req.Content),
	}
	if err := db.SystemTunnelTemplate.Create(rec); err != nil {
		return errors.New("创建失败")
	}
	return nil
}

func (s *service) Delete(claims jwt.Claims, code string) error {
	db, _, _ := repository.Get("")
	info, err := db.SystemTunnelTemplate.Where(db.SystemTunnelTemplate.Code.Eq(code), db.SystemTunnelTemplate.UserCode.Eq(claims.Code)).Delete()
	if err != nil {
		return errors.New("删除失败")
	}
	if info.RowsAffected == 0 {
		return errors.New("模板不存在")
	}
	return nil
}

type PageReq struct {
	bean.PageParam
	Kind string `json:"kind"`
}

func (s *service) Page(claims jwt.Claims, req PageReq) ([]*model.SystemTunnelTemplate, int64) {
	db, _, _ := repository.Get("")
	q := db.SystemTunnelTemplate.Where(db.SystemTunnelTemplate.UserCode.Eq(claims.Code)).Order(db.SystemTunnelTemplate.Id.Desc())
	if req.Kind != "" {
		q = q.Where(db.SystemTunnelTemplate.Kind.Eq(req.Kind))
	}
	list, total, _ := q.FindByPage(req.GetOffset(), req.GetLimit())
	return list, total
}
