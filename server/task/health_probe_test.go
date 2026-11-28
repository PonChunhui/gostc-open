package task

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"server/model"
)

// nodeForServer 构造一个指向 httptest 服务器的节点，使 probeHTTP 命中该服务器。
func nodeForServer(t *testing.T, srv *httptest.Server) *model.GostNode {
	t.Helper()
	addr := strings.TrimPrefix(srv.URL, "http://")
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("split host port: %v", err)
	}
	return &model.GostNode{Address: host + ":" + port, HttpPort: port}
}

func TestProbeHTTPRegistrationHint(t *testing.T) {
	cases := []struct {
		name      string
		status    int
		body      string
		wantReach int
		wantHint  int
	}{
		{"backend ok", 200, "hello", 1, 1},
		{"backend own 404", 404, "not found", 1, 1},
		{"frp 404 unregistered", 404, "The page you requested was not found. The server is powered by frp.", 2, 2},
		{"upstream 5xx registered", 502, "bad gateway", 2, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()
			node := nodeForServer(t, srv)
			reach, hint, _ := probeHTTP(node, "demo.example.com", time.Second*3)
			if reach != tc.wantReach || hint != tc.wantHint {
				t.Fatalf("got reach=%d hint=%d, want reach=%d hint=%d", reach, hint, tc.wantReach, tc.wantHint)
			}
		})
	}
}

func TestProbeHTTPConnErrorNoHint(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	node := nodeForServer(t, srv)
	srv.Close() // 关闭后连接失败，注册态不应被臆断
	reach, hint, _ := probeHTTP(node, "demo.example.com", time.Second*2)
	if reach != 2 || hint != 0 {
		t.Fatalf("conn error should give reach=2 hint=0, got reach=%d hint=%d", reach, hint)
	}
}

func TestProbeHTTPMissingTargetNoProbe(t *testing.T) {
	node := &model.GostNode{Address: "127.0.0.1:7000", HttpPort: ""}
	reach, hint, _ := probeHTTP(node, "demo.example.com", time.Second)
	if reach != 0 || hint != 0 {
		t.Fatalf("missing http port should skip probe, got reach=%d hint=%d", reach, hint)
	}
}
