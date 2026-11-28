package service

import (
	"errors"
	"time"

	"server/model"
	"server/pkg/bean"
	"server/pkg/jwt"
	"server/repository"
	"server/service/common/apitoken"
)

type CreateReq struct {
	Name     string `json:"name" binding:"required"`
	ExpDays  int    `json:"expDays"` // 0 表示永不过期
}

type CreateResp struct {
	Code  string `json:"code"`
	Token string `json:"token"` // 明文仅返回一次
}

func (s *service) Create(claims jwt.Claims, req CreateReq) (CreateResp, error) {
	token, err := apitoken.Generate()
	if err != nil {
		return CreateResp{}, errors.New("生成Token失败")
	}
	var expAt int64
	if req.ExpDays > 0 {
		expAt = time.Now().Add(time.Duration(req.ExpDays) * 24 * time.Hour).Unix()
	}
	db, _, _ := repository.Get("")
	rec := &model.SystemApiToken{
		UserCode:    claims.Code,
		Name:        req.Name,
		TokenHash:   apitoken.Hash(token),
		TokenPrefix: apitoken.Prefix(token),
		ExpAt:       expAt,
	}
	if err := db.SystemApiToken.Create(rec); err != nil {
		return CreateResp{}, errors.New("创建失败")
	}
	return CreateResp{Code: rec.Code, Token: token}, nil
}

func (s *service) Delete(claims jwt.Claims, code string) error {
	db, _, _ := repository.Get("")
	info, err := db.SystemApiToken.Where(db.SystemApiToken.Code.Eq(code), db.SystemApiToken.UserCode.Eq(claims.Code)).Delete()
	if err != nil {
		return errors.New("删除失败")
	}
	if info.RowsAffected == 0 {
		return errors.New("Token不存在")
	}
	return nil
}

type PageReq struct {
	bean.PageParam
}

type Item struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	TokenPrefix string `json:"tokenPrefix"`
	ExpAt      int64  `json:"expAt"`
	LastUsedAt int64  `json:"lastUsedAt"`
	CreatedAt  int64  `json:"createdAt"`
}

func (s *service) Page(claims jwt.Claims, req PageReq) ([]Item, int64) {
	db, _, _ := repository.Get("")
	list, total, _ := db.SystemApiToken.Where(db.SystemApiToken.UserCode.Eq(claims.Code)).Order(db.SystemApiToken.Id.Desc()).FindByPage(req.GetOffset(), req.GetLimit())
	items := make([]Item, 0, len(list))
	for _, v := range list {
		items = append(items, Item{
			Code:        v.Code,
			Name:        v.Name,
			TokenPrefix: v.TokenPrefix,
			ExpAt:       v.ExpAt,
			LastUsedAt:  v.LastUsedAt,
			CreatedAt:   v.CreatedAt.Unix(),
		})
	}
	return items, total
}
