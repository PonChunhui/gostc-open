package event

import (
	"github.com/lesismal/arpc"
	"gostc-sub/internal/common"
)

func LoggerHandle(client *arpc.Client) {
	client.Handler.Handle("logger", func(c *arpc.Context) {
		_ = c.Write(common.Logger.GetLogs())
	})
}