package system_cdk

import (
	"github.com/gin-gonic/gin"
	"server/pkg/bean"
	"server/service/admin/system_cdk"
)

var svr = service.Service

func BatchCreate(c *gin.Context) {
	var req service.BatchCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		bean.Response.Param(c, err)
		return
	}
	if err := svr.BatchCreate(req); err != nil {
		bean.Response.Fail(c, err.Error())
		return
	}
	bean.Response.Ok(c)
}

func BatchPage(c *gin.Context) {
	var req service.BatchPageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		bean.Response.Param(c, err)
		return
	}
	list, total := svr.BatchPage(req)
	bean.Response.OkData(c, bean.NewPage(list, total))
}

func BatchDelete(c *gin.Context) {
	var req service.BatchDeleteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		bean.Response.Param(c, err)
		return
	}
	if err := svr.BatchDelete(req); err != nil {
		bean.Response.Fail(c, err.Error())
		return
	}
	bean.Response.Ok(c)
}

func CardPage(c *gin.Context) {
	var req service.CardPageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		bean.Response.Param(c, err)
		return
	}
	list, total := svr.CardPage(req)
	bean.Response.OkData(c, bean.NewPage(list, total))
}

func CardExport(c *gin.Context) {
	var req service.CardExportReq
	if err := c.ShouldBindJSON(&req); err != nil {
		bean.Response.Param(c, err)
		return
	}
	secrets, err := svr.CardExport(req)
	if err != nil {
		bean.Response.Fail(c, err.Error())
		return
	}
	bean.Response.OkData(c, secrets)
}

func CardBan(c *gin.Context) {
	var req service.CardBanReq
	if err := c.ShouldBindJSON(&req); err != nil {
		bean.Response.Param(c, err)
		return
	}
	if err := svr.CardBan(req); err != nil {
		bean.Response.Fail(c, err.Error())
		return
	}
	bean.Response.Ok(c)
}