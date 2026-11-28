package normal

import (
	"github.com/gin-gonic/gin"
	"server/controller/normal/metric"
	"server/global"
	"server/router/middleware"
)

func InitMetric(group *gin.RouterGroup) {
	g := group.Group("metric", middleware.Auth(global.Jwt))
	g.POST("realtime", metric.Realtime)
	g.POST("history", metric.History)
}
