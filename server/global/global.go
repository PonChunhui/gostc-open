package global

import (
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"server/configs"
	"server/pkg/jwt"
	"server/pkg/memory"
	"server/pkg/orm"
)

var (
	Logger *zap.Logger
	Jwt    *jwt.Tool
	Cache  memory.Interface
	Config configs.Config
	DB     orm.Interface
	Cron   *cron.Cron

	// HealthProbeTrigger 由 task 包注入，触发一次（防抖的）即时健康探测。
	// 用于节点重连/面板重启后立即重建隧道注册态，而不必等待客户端重连。
	// 放在 global 以避免 router/rpc → task → bootstrap → router 的导入环。
	HealthProbeTrigger func()
)
