package download

import (
	"github.com/gin-gonic/gin"
	"server/pkg/bean"
	"server/service/public/download"
)

var svr = service.Service

func List(c *gin.Context) {
	bean.Response.OkData(c, svr.List())
}

func File(c *gin.Context) {
	name := c.Query("name")
	path, err := svr.File(name)
	if err != nil {
		bean.Response.Fail(c, err.Error())
		return
	}
	c.FileAttachment(path, name)
}
