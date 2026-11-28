package metric

import (
	"github.com/gin-gonic/gin"
	"server/pkg/bean"
	"server/service/admin/metric"
	normalmetric "server/service/normal/metric"
)

var svr = service.Service

func Realtime(c *gin.Context) {
	var req normalmetric.RealtimeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		bean.Response.Param(c, err)
		return
	}
	bean.Response.OkData(c, svr.Realtime(req))
}

func History(c *gin.Context) {
	var req normalmetric.HistoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		bean.Response.Param(c, err)
		return
	}
	bean.Response.OkData(c, svr.History(req))
}

func Top(c *gin.Context) {
	var req service.TopReq
	if err := c.ShouldBindJSON(&req); err != nil {
		bean.Response.Param(c, err)
		return
	}
	bean.Response.OkData(c, svr.Top(req))
}

func Objects(c *gin.Context) {
	var req service.ObjectsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		bean.Response.Param(c, err)
		return
	}
	bean.Response.OkData(c, svr.Objects(req))
}
