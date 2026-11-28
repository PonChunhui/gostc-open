package openapi

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

var routes gin.RoutesInfo

// SetRoutes 由 bootstrap 在路由注册完成后注入，用于生成 OpenAPI 文档。
func SetRoutes(r gin.RoutesInfo) {
	routes = r
}

func convertPath(p string) string {
	segs := strings.Split(p, "/")
	for i, s := range segs {
		if strings.HasPrefix(s, ":") {
			segs[i] = "{" + s[1:] + "}"
		} else if strings.HasPrefix(s, "*") {
			segs[i] = "{" + s[1:] + "}"
		}
	}
	return strings.Join(segs, "/")
}

// Spec 输出由实际注册路由生成的 OpenAPI 3.0 文档。
func Spec(c *gin.Context) {
	paths := map[string]any{}
	for _, r := range routes {
		if !strings.HasPrefix(r.Path, "/api/v1") {
			continue
		}
		p := convertPath(r.Path)
		method := strings.ToLower(r.Method)
		trimmed := strings.TrimPrefix(r.Path, "/api/v1/")
		tag := strings.SplitN(trimmed, "/", 2)[0]
		op := map[string]any{
			"tags":        []string{tag},
			"summary":     r.Handler,
			"operationId": method + "_" + strings.ReplaceAll(strings.Trim(p, "/"), "/", "_"),
			"responses": map[string]any{
				"200": map[string]any{"description": "OK"},
			},
		}
		if method != "get" {
			op["requestBody"] = map[string]any{
				"content": map[string]any{
					"application/json": map[string]any{"schema": map[string]any{"type": "object"}},
				},
			}
		}
		item, ok := paths[p].(map[string]any)
		if !ok {
			item = map[string]any{}
			paths[p] = item
		}
		item[method] = op
	}
	spec := map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":       "GOSTC Open API",
			"description": "内网穿透平台开放接口。鉴权：请求头 token 携带登录 JWT 或长期 API Token（gostc_ 前缀）。",
			"version":     "1.0.0",
		},
		"servers": []any{map[string]any{"url": "/"}},
		"components": map[string]any{
			"securitySchemes": map[string]any{
				"token": map[string]any{"type": "apiKey", "in": "header", "name": "token"},
			},
		},
		"security": []any{map[string]any{"token": []string{}}},
		"paths":    paths,
	}
	c.JSON(http.StatusOK, spec)
}

// Docs 提供 Swagger UI 在线浏览页面。
func Docs(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`<!doctype html><html><head><meta charset="utf-8">
<title>GOSTC API 文档</title>
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css">
</head><body><div id="swagger"></div>
<script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
<script>
window.onload=()=>{SwaggerUIBundle({url:'/api/v1/public/openapi/spec',dom_id:'#swagger',deepLinking:true});};
</script></body></html>`))
}
