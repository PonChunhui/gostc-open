package service

import (
	"errors"
	"github.com/SianHH/frp-package/server/metrics"
	"github.com/lesismal/arpc"
	"sync/atomic"
	"time"
)

type Metrics struct {
	client          *arpc.Client
	clientCount     atomic.Int64
	tunnelCount     atomic.Int64
	connectionCount atomic.Int64
}

func InitMetrics(client *arpc.Client) error {
	m := &Metrics{
		client: client,
	}
	metrics.Server = m
	go m.reportLoad()
	return nil
}

func (m *Metrics) NewClient() {
	m.clientCount.Add(1)
}

func (m *Metrics) CloseClient() {
	m.clientCount.Add(-1)
}

func (m *Metrics) NewProxy(name string, proxyType string) {
	m.tunnelCount.Add(1)
}

func (m *Metrics) CloseProxy(name string, proxyType string) {
	m.tunnelCount.Add(-1)
}

func (m *Metrics) OpenConnection(name string, proxyType string) {
	m.connectionCount.Add(1)
}

func (m *Metrics) CloseConnection(name string, proxyType string) {
	m.connectionCount.Add(-1)
}

// NodeLoadData 节点实时负载上报数据
type NodeLoadData struct {
	ClientCount     int64 `json:"clientCount"`
	TunnelCount     int64 `json:"tunnelCount"`
	ConnectionCount int64 `json:"connectionCount"`
}

// reportLoad 定时上报节点负载，供服务端展示隧道使用情况
func (m *Metrics) reportLoad() {
	for {
		time.Sleep(time.Second * 10)
		if err := m.client.CheckState(); err != nil {
			if !errors.Is(err, arpc.ErrClientReconnecting) {
				return
			}
			continue
		}
		_ = m.client.CallAsync("rpc/node/load", NodeLoadData{
			ClientCount:     max(m.clientCount.Load(), 0),
			TunnelCount:     max(m.tunnelCount.Load(), 0),
			ConnectionCount: max(m.connectionCount.Load(), 0),
		}, func(c *arpc.Context, err error) {
		}, time.Second*5)
	}
}

type TrafficData struct {
	Name      string
	ProxyType string
	Total     int64
}

func (m *Metrics) AddTrafficIn(name string, proxyType string, trafficBytes int64) {
	_ = m.client.CallAsync("rpc/metrics/input", TrafficData{
		Name:      name,
		ProxyType: proxyType,
		Total:     trafficBytes,
	}, func(c *arpc.Context, err error) {
	}, time.Second*10)
}

func (m *Metrics) AddTrafficOut(name string, proxyType string, trafficBytes int64) {
	_ = m.client.CallAsync("rpc/metrics/output", TrafficData{
		Name:      name,
		ProxyType: proxyType,
		Total:     trafficBytes,
	}, func(c *arpc.Context, err error) {
	}, time.Second*10)
}
