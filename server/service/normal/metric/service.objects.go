package service

import (
	"server/pkg/jwt"
	"server/repository"
	"sort"
)

// ObjectItem 可供指标查询选择的对象（编码+名称）。
type ObjectItem struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// ObjectsReq 按对象类型列出可选对象。
type ObjectsReq struct {
	OriginKind int `json:"originKind"` // 1节点 2客户端 3隧道 4用户
}

// Objects 返回当前用户可选的对象，供实时/历史曲线选择，避免手填编码。
func (s *service) Objects(claims jwt.Claims, req ObjectsReq) []ObjectItem {
	db, _, _ := repository.Get("")
	items := make([]ObjectItem, 0)
	switch req.OriginKind {
	case 1:
		nodes, _ := db.GostNode.Find()
		for _, n := range nodes {
			items = append(items, ObjectItem{Code: n.Code, Name: n.Name})
		}
	case 2:
		clients, _ := db.GostClient.Where(db.GostClient.UserCode.Eq(claims.Code)).Find()
		for _, c := range clients {
			items = append(items, ObjectItem{Code: c.Code, Name: c.Name})
		}
	case 3:
		hosts, _ := db.GostClientHost.Where(db.GostClientHost.UserCode.Eq(claims.Code)).Find()
		for _, h := range hosts {
			items = append(items, ObjectItem{Code: h.Code, Name: h.Name})
		}
		forwards, _ := db.GostClientForward.Where(db.GostClientForward.UserCode.Eq(claims.Code)).Find()
		for _, f := range forwards {
			items = append(items, ObjectItem{Code: f.Code, Name: f.Name})
		}
		proxies, _ := db.GostClientProxy.Where(db.GostClientProxy.UserCode.Eq(claims.Code)).Find()
		for _, p := range proxies {
			items = append(items, ObjectItem{Code: p.Code, Name: p.Name})
		}
		tunnels, _ := db.GostClientTunnel.Where(db.GostClientTunnel.UserCode.Eq(claims.Code)).Find()
		for _, t := range tunnels {
			items = append(items, ObjectItem{Code: t.Code, Name: t.Name})
		}
		p2ps, _ := db.GostClientP2P.Where(db.GostClientP2P.UserCode.Eq(claims.Code)).Find()
		for _, p := range p2ps {
			items = append(items, ObjectItem{Code: p.Code, Name: p.Name})
		}
	case 4:
		items = append(items, ObjectItem{Code: claims.Code, Name: "我的账号"})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
	return items
}
