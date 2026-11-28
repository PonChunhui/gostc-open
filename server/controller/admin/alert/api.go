package alert

import (
	"github.com/gin-gonic/gin"
	"server/pkg/bean"
	"server/service/admin/alert"
)

var svr = service.Service

func RulePage(c *gin.Context) {
	var req service.RulePageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		bean.Response.Param(c, err)
		return
	}
	list, total := svr.PageRule(req)
	bean.Response.OkData(c, bean.NewPage(list, total))
}

func RuleDelete(c *gin.Context) {
	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bean.Response.Param(c, err)
		return
	}
	if err := svr.DeleteRule(req.Code); err != nil {
		bean.Response.Fail(c, err.Error())
		return
	}
	bean.Response.Ok(c)
}

func LogPage(c *gin.Context) {
	var req service.LogPageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		bean.Response.Param(c, err)
		return
	}
	list, total := svr.PageLog(req)
	bean.Response.OkData(c, bean.NewPage(list, total))
}
