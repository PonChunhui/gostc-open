package public

import (
	"github.com/gin-gonic/gin"
	"server/controller/public/client_version"
)

func InitClientVersion(group *gin.RouterGroup) {
	g := group.Group("client/version")
	g.GET("latest", client_version.Latest)
}
