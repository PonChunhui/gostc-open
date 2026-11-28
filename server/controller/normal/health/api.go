package health

import (
	"github.com/gin-gonic/gin"
	"server/pkg/bean"
	"server/router/middleware"
	"server/service/normal/health"
)

var svr = service.Service

func Page(c *gin.Context) {
	var req service.PageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		bean.Response.Param(c, err)
		return
	}
	list, total := svr.Page(middleware.GetClaims(c), req)
	bean.Response.OkData(c, bean.NewPage(list, total))
}

func Overview(c *gin.Context) {
	bean.Response.OkData(c, svr.Overview(middleware.GetClaims(c)))
}
