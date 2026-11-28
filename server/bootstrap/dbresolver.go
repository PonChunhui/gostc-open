package bootstrap

import (
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
	"server/global"
)

// enableDBResolver 按配置启用读写分离：写走主库（默认连接），读随机走从库。
func enableDBResolver() {
	cfg := global.Config.DBResolver
	if !cfg.Enable || len(cfg.Replicas) == 0 {
		return
	}
	if global.Config.DbType != "mysql" && global.Config.DbType != "Mysql" {
		global.Logger.Warn("读写分离仅支持 mysql，已忽略 db-resolver 配置")
		return
	}
	var replicas []gorm.Dialector
	for _, dsn := range cfg.Replicas {
		if dsn == "" {
			continue
		}
		replicas = append(replicas, mysql.Open(dsn))
	}
	if len(replicas) == 0 {
		return
	}
	err := global.DB.GetDB().Use(dbresolver.Register(dbresolver.Config{
		Replicas: replicas,
		Policy:   dbresolver.RandomPolicy{},
	}))
	if err != nil {
		global.Logger.Warn("启用读写分离失败", zap.Error(err))
		return
	}
	global.Logger.Info("启用读写分离", zap.Int("replicas", len(replicas)))
}
