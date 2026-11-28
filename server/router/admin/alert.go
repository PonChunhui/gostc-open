package admin

import (
	"github.com/gin-gonic/gin"
	"server/controller/admin/alert"
	"server/global"
	"server/router/middleware"
)

func InitAlert(group *gin.RouterGroup) {
	g := group.Group("alert", middleware.Auth(global.Jwt), middleware.AuthAdmin())
	g.POST("rule/page", alert.RulePage)
	g.POST("rule/delete", alert.RuleDelete)
	g.POST("log/page", alert.LogPage)
}
