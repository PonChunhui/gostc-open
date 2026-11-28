package event

import (
	"bytes"
	"encoding/json"
	"github.com/lesismal/arpc"
	"net/http"
)

func ServerDomainHandle(client *arpc.Client, proxyBaseUrl string) {
	client.Handler.Handle("server_domain_config", func(c *arpc.Context) {
		if proxyBaseUrl == "" {
			_ = c.Write("not support")
			return
		}
		var req ServerDomain
		if err := c.Bind(&req); err != nil {
			_ = c.Write(err.Error())
			return
		}
		if err := domainRequestDo(proxyBaseUrl, domainRequestData{
			Domain:     req.Domain,
			Target:     req.Target,
			Cert:       req.Cert,
			Key:        req.Key,
			ForceHttps: req.ForceHttps,
		}); err != nil {
			_ = c.Write(err)
			return
		}
		_ = c.Write("success")
	})
}

type domainRequestData struct {
	Domain     string `json:"domain"`
	Target     string `json:"target"`
	Cert       string `json:"cert"`
	Key        string `json:"key"`
	ForceHttps int    `json:"forceHttps"`
}

// ServerDomainRemoveHandle 处理源网关域名路由清理（故障转移/解绑/删证书）。
func ServerDomainRemoveHandle(client *arpc.Client, proxyBaseUrl string) {
	client.Handler.Handle("server_domain_remove", func(c *arpc.Context) {
		if proxyBaseUrl == "" {
			_ = c.Write("not support")
			return
		}
		var req ServerDomainRemove
		if err := c.Bind(&req); err != nil {
			_ = c.Write(err.Error())
			return
		}
		marshal, _ := json.Marshal(map[string]any{"domain": req.Domain, "purgeCert": req.PurgeCert})
		request, err := http.NewRequest(http.MethodPost, proxyBaseUrl+"/domain/delete", bytes.NewReader(marshal))
		if err != nil {
			_ = c.Write(err)
			return
		}
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			_ = c.Write(err)
			return
		}
		_ = response.Body.Close()
		_ = c.Write("success")
	})
}

type ServerDomainRemove struct {
	Domain    string `json:"domain"`
	PurgeCert bool   `json:"purgeCert"`
}

func domainRequestDo(baseUrl string, data domainRequestData) error {
	marshal, _ := json.Marshal(data)
	request, err := http.NewRequest(http.MethodPost, baseUrl+"/domain", bytes.NewReader(marshal))
	if err != nil {
		return err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	_ = response.Body.Close()
	return nil
}
