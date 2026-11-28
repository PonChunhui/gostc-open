package config

import (
	"github.com/gin-gonic/gin"
	"gostc-sub/pkg/bean"
	service "gostc-sub/webui/backend/service/config"
)

func UpdateQuery(c *gin.Context) {
	bean.Response.OkData(c, service.Service.UpdateQuery())
}

func UpdateSave(c *gin.Context) {
	var req service.UpdateSettingsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		bean.Response.Param(c, err)
		return
	}
	if err := service.Service.UpdateSave(req); err != nil {
		bean.Response.Fail(c, err.Error())
		return
	}
	bean.Response.Ok(c)
}

func UpdateCheck(c *gin.Context) {
	var req service.UpdateCheckReq
	_ = c.ShouldBindJSON(&req)
	info, err := service.Service.UpdateCheck(req)
	if err != nil {
		bean.Response.Fail(c, err.Error())
		return
	}
	bean.Response.OkData(c, info)
}

func UpdateApply(c *gin.Context) {
	var req service.UpdateCheckReq
	_ = c.ShouldBindJSON(&req)
	if err := service.Service.UpdateApply(req); err != nil {
		bean.Response.Fail(c, err.Error())
		return
	}
	bean.Response.Ok(c)
}
