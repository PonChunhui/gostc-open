package alert

import (
	"github.com/gin-gonic/gin"
	"server/pkg/bean"
	"server/router/middleware"
	"server/service/normal/alert"
)

var svr = service.Service

func RuleCreate(c *gin.Context) {
	var req service.RuleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		bean.Response.Param(c, err)
		return
	}
	if err := svr.CreateRule(middleware.GetClaims(c), req); err != nil {
		bean.Response.Fail(c, err.Error())
		return
	}
	bean.Response.Ok(c)
}

func RuleUpdate(c *gin.Context) {
	var req service.RuleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		bean.Response.Param(c, err)
		return
	}
	if err := svr.UpdateRule(middleware.GetClaims(c), req); err != nil {
		bean.Response.Fail(c, err.Error())
		return
	}
	bean.Response.Ok(c)
}

func RuleDelete(c *gin.Context) {
	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bean.Response.Param(c, err)
		return
	}
	if err := svr.DeleteRule(middleware.GetClaims(c), req.Code); err != nil {
		bean.Response.Fail(c, err.Error())
		return
	}
	bean.Response.Ok(c)
}

func RulePage(c *gin.Context) {
	var req service.RulePageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		bean.Response.Param(c, err)
		return
	}
	list, total := svr.PageRule(middleware.GetClaims(c), req)
	bean.Response.OkData(c, bean.NewPage(list, total))
}

func LogPage(c *gin.Context) {
	var req service.LogPageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		bean.Response.Param(c, err)
		return
	}
	list, total := svr.PageLog(middleware.GetClaims(c), req)
	bean.Response.OkData(c, bean.NewPage(list, total))
}
