package public

import (
	"github.com/gin-gonic/gin"
	"server/controller/public/pay"
)

func InitPay(group *gin.RouterGroup) {
	g := group.Group("pay")
	g.POST("notify/:provider", pay.Notify)
	g.GET("sandbox/page", pay.SandboxPage)
	g.POST("sandbox/confirm", pay.SandboxConfirm)
}
