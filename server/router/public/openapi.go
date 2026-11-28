package public

import (
	"github.com/gin-gonic/gin"
	"server/controller/public/openapi"
)

func InitOpenApi(group *gin.RouterGroup) {
	g := group.Group("openapi")
	g.GET("spec", openapi.Spec)
	g.GET("docs", openapi.Docs)
}
