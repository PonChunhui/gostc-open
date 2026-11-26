package public

import (
	"github.com/gin-gonic/gin"
	"server/controller/public/download"
)

func InitDownload(group *gin.RouterGroup) {
	g := group.Group("download")
	g.POST("list", download.List)
	g.GET("file", download.File)
}
