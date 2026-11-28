package service

import (
	"errors"

	"gorm.io/gen"
	"server/model"
	"server/pkg/bean"
	"server/repository"
)

type service struct{}

var Service = &service{}

type VersionReq struct {
	Code        string `json:"code"`
	Version     string `json:"version" binding:"required"`
	Platform    string `json:"platform" binding:"required"`
	Arch        string `json:"arch"`
	DownloadURL string `json:"downloadUrl" binding:"required"`
	Sha256      string `json:"sha256"`
	Signature   string `json:"signature"`
	Channel     string `json:"channel"`
	GrayPercent int    `json:"grayPercent"`
	Enable      int    `json:"enable"`
	ReleaseNote string `json:"releaseNote"`
}

func validChannel(c string) bool {
	return c == model.CLIENT_VERSION_CHANNEL_STABLE || c == model.CLIENT_VERSION_CHANNEL_BETA
}

func (s *service) Create(req VersionReq) error {
	if req.Channel == "" {
		req.Channel = model.CLIENT_VERSION_CHANNEL_STABLE
	}
	if !validChannel(req.Channel) {
		return errors.New("通道无效")
	}
	if req.GrayPercent < 0 || req.GrayPercent > 100 {
		return errors.New("灰度百分比需在0-100")
	}
	db, _, _ := repository.Get("")
	v := &model.SystemClientVersion{
		Version: req.Version, Platform: req.Platform, Arch: req.Arch,
		DownloadURL: req.DownloadURL, Sha256: req.Sha256, Signature: req.Signature,
		Channel: req.Channel, GrayPercent: req.GrayPercent, ReleaseNote: req.ReleaseNote,
		Enable: 1,
	}
	if req.Enable == 2 {
		v.Enable = 2
	}
	if err := db.SystemClientVersion.Create(v); err != nil {
		return errors.New("创建版本失败")
	}
	return nil
}

func (s *service) Update(req VersionReq) error {
	if req.Code == "" {
		return errors.New("缺少编号")
	}
	db, _, _ := repository.Get("")
	v, err := db.SystemClientVersion.Where(db.SystemClientVersion.Code.Eq(req.Code)).First()
	if err != nil {
		return errors.New("版本不存在")
	}
	if req.Channel != "" && !validChannel(req.Channel) {
		return errors.New("通道无效")
	}
	if req.GrayPercent < 0 || req.GrayPercent > 100 {
		return errors.New("灰度百分比需在0-100")
	}
	v.Version = req.Version
	v.Platform = req.Platform
	v.Arch = req.Arch
	v.DownloadURL = req.DownloadURL
	v.Sha256 = req.Sha256
	v.Signature = req.Signature
	if req.Channel != "" {
		v.Channel = req.Channel
	}
	v.GrayPercent = req.GrayPercent
	v.ReleaseNote = req.ReleaseNote
	if req.Enable == 1 || req.Enable == 2 {
		v.Enable = req.Enable
	}
	if err := db.SystemClientVersion.Save(v); err != nil {
		return errors.New("更新版本失败")
	}
	return nil
}

func (s *service) Delete(code string) error {
	db, _, _ := repository.Get("")
	if _, err := db.SystemClientVersion.Where(db.SystemClientVersion.Code.Eq(code)).Delete(); err != nil {
		return errors.New("删除版本失败")
	}
	return nil
}

type PageReq struct {
	bean.PageParam
	Platform string `json:"platform"`
	Channel  string `json:"channel"`
}

func (s *service) Page(req PageReq) ([]*model.SystemClientVersion, int64) {
	db, _, _ := repository.Get("")
	q := db.SystemClientVersion.Order(db.SystemClientVersion.Id.Desc())
	var where []gen.Condition
	if req.Platform != "" {
		where = append(where, db.SystemClientVersion.Platform.Eq(req.Platform))
	}
	if req.Channel != "" {
		where = append(where, db.SystemClientVersion.Channel.Eq(req.Channel))
	}
	list, total, _ := q.Where(where...).FindByPage(req.GetOffset(), req.GetLimit())
	return list, total
}
