package normal

import (
	"github.com/gin-gonic/gin"
	"server/controller/normal/alert"
	"server/global"
	"server/router/middleware"
)

func InitAlert(group *gin.RouterGroup) {
	g := group.Group("alert", middleware.Auth(global.Jwt))
	g.POST("rule/create", alert.RuleCreate)
	g.POST("rule/update", alert.RuleUpdate)
	g.POST("rule/delete", alert.RuleDelete)
	g.POST("rule/page", alert.RulePage)
	g.POST("log/page", alert.LogPage)
}
