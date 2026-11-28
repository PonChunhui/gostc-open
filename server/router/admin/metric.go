package admin

import (
	"github.com/gin-gonic/gin"
	"server/controller/admin/metric"
	"server/global"
	"server/router/middleware"
)

func InitMetric(group *gin.RouterGroup) {
	g := group.Group("metric", middleware.Auth(global.Jwt), middleware.AuthAdmin())
	g.POST("realtime", metric.Realtime)
	g.POST("history", metric.History)
	g.POST("top", metric.Top)
}
