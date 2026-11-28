package client_version

import (
	"hash/fnv"

	"github.com/gin-gonic/gin"
	"server/model"
	"server/pkg/bean"
	"server/repository"
)

type LatestResp struct {
	Version     string `json:"version"`
	DownloadURL string `json:"downloadUrl"`
	Sha256      string `json:"sha256"`
	Signature   string `json:"signature"`
	Channel     string `json:"channel"`
	ReleaseNote string `json:"releaseNote"`
}

// Latest 客户端检查更新：按平台/架构/通道返回最新启用版本，灰度按 uid 哈希分桶。
// GET /api/v1/public/client/version/latest?platform=&arch=&channel=&uid=&current=
func Latest(c *gin.Context) {
	platform := c.Query("platform")
	arch := c.Query("arch")
	channel := c.Query("channel")
	uid := c.Query("uid")
	current := c.Query("current")
	if channel == "" {
		channel = model.CLIENT_VERSION_CHANNEL_STABLE
	}
	if platform == "" {
		bean.Response.Fail(c, "缺少platform")
		return
	}

	db, _, _ := repository.Get("")
	q := db.SystemClientVersion.Where(
		db.SystemClientVersion.Enable.Eq(1),
		db.SystemClientVersion.Platform.Eq(platform),
		db.SystemClientVersion.Channel.Eq(channel),
	).Order(db.SystemClientVersion.Id.Desc())
	if arch != "" {
		q = q.Where(db.SystemClientVersion.Arch.Eq(arch))
	}
	list, err := q.Find()
	if err != nil || len(list) == 0 {
		bean.Response.OkData(c, nil)
		return
	}

	for _, v := range list {
		if v.Version == current {
			continue // 已是该版本
		}
		if !inGray(v.GrayPercent, uid) {
			continue // 未命中灰度
		}
		bean.Response.OkData(c, LatestResp{
			Version: v.Version, DownloadURL: v.DownloadURL, Sha256: v.Sha256,
			Signature: v.Signature, Channel: v.Channel, ReleaseNote: v.ReleaseNote,
		})
		return
	}
	bean.Response.OkData(c, nil)
}

func inGray(percent int, uid string) bool {
	if percent >= 100 || uid == "" {
		return true
	}
	if percent <= 0 {
		return false
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(uid))
	return int(h.Sum32()%100) < percent
}
