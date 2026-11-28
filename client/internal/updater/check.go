package updater

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"time"
)

// Info 面板返回的最新版本信息（对应服务端 LatestResp）。
type Info struct {
	Version     string `json:"version"`
	DownloadURL string `json:"downloadUrl"`
	Sha256      string `json:"sha256"`
	Signature   string `json:"signature"`
	Channel     string `json:"channel"`
	ReleaseNote string `json:"releaseNote"`
}

type latestResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data *Info  `json:"data"`
}

var checkClient = &http.Client{
	Timeout: time.Second * 20,
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	},
}

// CheckLatest 查询面板最新版本。无更新（data 为空/版本不高于当前）返回 (nil, nil)。
// base 为面板基址（如 https://gostc.shop）；uid 用于灰度分桶（客户端/节点 key）。
func CheckLatest(base, channel, uid, current string) (*Info, error) {
	if base == "" {
		return nil, fmt.Errorf("面板地址为空")
	}
	if channel == "" {
		channel = ChannelStable
	}
	q := url.Values{}
	q.Set("platform", runtime.GOOS)
	q.Set("arch", runtime.GOARCH)
	q.Set("channel", channel)
	q.Set("uid", uid)
	q.Set("current", current)
	endpoint := base + "/api/v1/public/client/version/latest?" + q.Encode()

	resp, err := checkClient.Get(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("检查更新返回状态 %d", resp.StatusCode)
	}
	var r latestResp
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	if r.Code != 0 {
		return nil, fmt.Errorf("检查更新失败: %s", r.Msg)
	}
	if r.Data == nil || r.Data.Version == "" || r.Data.DownloadURL == "" {
		return nil, nil
	}
	if r.Data.Version == current || Compare(r.Data.Version, current) <= 0 {
		return nil, nil // 已是最新或更高，防降级
	}
	return r.Data, nil
}
