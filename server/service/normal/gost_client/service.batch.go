package service

import (
	"errors"

	"server/pkg/jwt"
	forwardservice "server/service/normal/gost_client_forward"
	hostservice "server/service/normal/gost_client_host"
)

type BatchReq struct {
	Kind   string   `json:"kind" binding:"required"`   // host / forward / client
	Action string   `json:"action" binding:"required"` // enable / disable / delete
	Codes  []string `json:"codes" binding:"required"`
}

type BatchResult struct {
	Code string `json:"code"`
	OK   bool   `json:"ok"`
	Msg  string `json:"msg"`
}

type BatchResp struct {
	Success int           `json:"success"`
	Failed  int           `json:"failed"`
	Results []BatchResult `json:"results"`
}

// Batch 批量启停/删除隧道或客户端，逐项执行并汇总结果（单项失败不影响其余）。
func (service *service) Batch(claims jwt.Claims, req BatchReq) (BatchResp, error) {
	if len(req.Codes) == 0 {
		return BatchResp{}, errors.New("请选择至少一项")
	}
	if len(req.Codes) > 200 {
		return BatchResp{}, errors.New("单次批量不超过200项")
	}
	var enable int
	switch req.Action {
	case "enable":
		enable = 1
	case "disable":
		enable = 2
	case "delete":
	default:
		return BatchResp{}, errors.New("操作类型无效")
	}

	resp := BatchResp{Results: make([]BatchResult, 0, len(req.Codes))}
	for _, code := range req.Codes {
		err := batchOne(claims, req.Kind, req.Action, enable, code)
		r := BatchResult{Code: code, OK: err == nil}
		if err != nil {
			r.Msg = err.Error()
			resp.Failed++
		} else {
			resp.Success++
		}
		resp.Results = append(resp.Results, r)
	}
	return resp, nil
}

func batchOne(claims jwt.Claims, kind, action string, enable int, code string) error {
	switch kind {
	case "host":
		if action == "delete" {
			return hostservice.Service.Delete(claims, hostservice.DeleteReq{Code: code})
		}
		return hostservice.Service.Enable(claims, hostservice.EnableReq{Code: code, Enable: enable})
	case "forward":
		if action == "delete" {
			return forwardservice.Service.Delete(claims, forwardservice.DeleteReq{Code: code})
		}
		return forwardservice.Service.Enable(claims, forwardservice.EnableReq{Code: code, Enable: enable})
	case "client":
		if action != "delete" {
			return errors.New("客户端仅支持批量删除")
		}
		return Service.Delete(claims, DeleteReq{Code: code})
	default:
		return errors.New("类型无效")
	}
}
