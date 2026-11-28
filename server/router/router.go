package router

import (
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"server/bootstrap"
	"server/global"
	"server/pkg/logger"
	"server/pkg/middleware"
	"server/pkg/rpc_protocol/websocket"
	routemw "server/router/middleware"
	"server/router/admin"
	"server/router/auth"
	"server/router/normal"
	"server/router/open"
	"server/router/public"
	"server/router/rpc"
	"strings"
)

func init() {
	bootstrap.Route = func(engine *gin.Engine) {
		// 使用GZIP压缩数据
		engine.Use(gzip.Gzip(gzip.DefaultCompression))

		// 记录HTTP请求日志
		httpLog := logger.NewLogger(logger.Option{
			To:         []string{global.BASE_PATH + "/data/httpLog/access.log"},
			Level:      "debug",
			MaxSize:    100,
			MaxAge:     30,
			MaxBackups: 10,
			Compress:   true,
		})
		engine.Use(middleware.Logger(httpLog, global.MODE == "dev", func(c *gin.Context) bool {
			return strings.Contains(c.Request.RequestURI, "api/v")
		}))

		// 审计日志：记录管理/用户侧写操作
		engine.Use(routemw.Audit())

		_ = InitStatic(engine)

		adminGroup := engine.Group("api/v1/admin")
		admin.InitGostClient(adminGroup)
		admin.InitGostClientLogger(adminGroup)
		admin.InitGostClientForward(adminGroup)
		admin.InitGostClientHost(adminGroup)
		admin.InitGostClientTunnel(adminGroup)
		admin.InitGostClientProxy(adminGroup)
		admin.InitGostClientP2P(adminGroup)
		admin.InitGostNodeConfig(adminGroup)
		admin.InitGostNode(adminGroup)
		admin.InitGostNodeBind(adminGroup)
		admin.InitGostNodeRule(adminGroup)
		admin.InitGostNodeGroup(adminGroup)
		admin.InitSystemUser(adminGroup)
		admin.InitSystemUserGroup(adminGroup)
		admin.InitSystemCdk(adminGroup)
		admin.InitSystemNotice(adminGroup)
		admin.InitSystemAuditLog(adminGroup)
		admin.InitSystemConfig(adminGroup)
		admin.InitDashboard(adminGroup)
		admin.InitClientVersion(adminGroup)
		admin.InitHealth(adminGroup)
		admin.InitAlert(adminGroup)
		admin.InitMetric(adminGroup)
		admin.InitFailover(adminGroup)

		publicGroup := engine.Group("api/v1/public")
		public.InitSystemConfig(publicGroup)
		public.InitFrp(publicGroup)
		public.InitGost(publicGroup)
		public.InitDownload(publicGroup)
		public.InitClientVersion(publicGroup)
		public.InitPay(publicGroup)
		public.InitOpenApi(publicGroup)

		authGroup := engine.Group("api/v1/auth")
		auth.InitAuth(authGroup)

		normalGroup := engine.Group("api/v1/normal")
		normal.InitGost(normalGroup)
		normal.InitGostClient(normalGroup)
		normal.InitGostClientLogger(normalGroup)
		normal.InitGostClientForward(normalGroup)
		normal.InitGostClientHost(normalGroup)
		normal.InitGostClientTunnel(normalGroup)
		normal.InitGostClientProxy(normalGroup)
		normal.InitGostClientP2P(normalGroup)
		normal.InitGostNode(normalGroup)
		normal.InitGostObs(normalGroup)
		normal.InitSystemNotice(normalGroup)
		normal.InitDashboard(normalGroup)
		normal.InitFrpClientCfg(normalGroup)
		normal.InitAlert(normalGroup)
		normal.InitHealth(normalGroup)
		normal.InitMetric(normalGroup)
		normal.InitFailover(normalGroup)
		normal.InitCert(normalGroup)
		normal.InitPay(normalGroup)
		normal.InitApiToken(normalGroup)
		normal.InitTemplate(normalGroup)

		openGroup := engine.Group("api/v1/open")
		open.InitAuth(openGroup.Group("auth"))

		// RPC Server
		ln, err := websocket.Listen(global.Config.Address, nil)
		if err != nil {
			panic("rpc server listen fail,err" + err.Error())
		}
		rpc.InitFrp(engine, ln)
	}
}
