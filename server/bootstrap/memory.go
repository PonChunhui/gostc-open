package bootstrap

import (
	"server/global"
	"server/pkg/memory"
	"strings"
	"time"

	"go.uber.org/zap"
)

func InitMemory() {
	cfg := global.Config.Cache
	if strings.EqualFold(cfg.Type, "redis") {
		r, err := memory.NewRedis(cfg.Redis.Addr, cfg.Redis.Pwd, cfg.Redis.DB, cfg.Redis.Prefix)
		if err == nil {
			global.Cache = r
			global.Logger.Info("缓存后端: redis", zap.String("addr", cfg.Redis.Addr))
			return
		}
		global.Logger.Warn("redis 不可用，降级为本地缓存", zap.Error(err))
	}
	global.Cache = memory.NewMemory(global.BASE_PATH+"/data/cache.db", time.Minute*10)
	releaseFunc = append(releaseFunc, func() {
		global.Cache.Sync()
	})
}
