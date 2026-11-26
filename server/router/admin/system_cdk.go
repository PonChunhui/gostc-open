package admin

import (
	"github.com/gin-gonic/gin"
	"server/controller/admin/system_cdk"
	"server/global"
	"server/router/middleware"
)

func InitSystemCdk(group *gin.RouterGroup) {
	g := group.Group("system/cdk", middleware.Auth(global.Jwt), middleware.AuthAdmin())
	g.POST("batch/create", system_cdk.BatchCreate)
	g.POST("batch/page", system_cdk.BatchPage)
	g.POST("batch/delete", system_cdk.BatchDelete)
	g.POST("card/page", system_cdk.CardPage)
	g.POST("card/export", system_cdk.CardExport)
	g.POST("card/ban", system_cdk.CardBan)
}