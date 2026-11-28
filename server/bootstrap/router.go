package bootstrap

import (
	"github.com/gin-gonic/gin"
	"server/controller/public/openapi"
	"server/global"
)

var engine *gin.Engine

var Route func(e *gin.Engine)

func InitRouter() {
	if global.MODE == "prod" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
	engine = gin.Default()
	
	if Route != nil {
		Route(engine)
	}
	openapi.SetRoutes(engine.Routes())
	global.Logger.Info("init router finish")
}
