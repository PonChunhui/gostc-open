package node_select

import (
	"math/rand"
	"server/model"
	"testing"
)

func TestPickByStrategy(t *testing.T) {
	nodes := []model.GostNode{
		{Base: model.Base{Code: "a"}, Priority: 5, Weight: 1},
		{Base: model.Base{Code: "b"}, Priority: 1, Weight: 1},
		{Base: model.Base{Code: "c"}, Priority: 3, Weight: 1},
	}
	rnd := rand.New(rand.NewSource(1))

	if got := PickByStrategy(nil, model.GOST_NODE_GROUP_STRATEGY_PRIORITY, rnd); got != nil {
		t.Fatalf("empty candidates should return nil, got %v", got)
	}

	// priority：选 Priority 最小者 b
	if got := PickByStrategy(nodes, model.GOST_NODE_GROUP_STRATEGY_PRIORITY, rnd); got == nil || got.Code != "b" {
		t.Fatalf("priority should pick b, got %v", got)
	}

	// random / weight：结果须在候选集合内
	for _, strategy := range []string{model.GOST_NODE_GROUP_STRATEGY_RANDOM, model.GOST_NODE_GROUP_STRATEGY_WEIGHT} {
		for i := 0; i < 20; i++ {
			got := PickByStrategy(nodes, strategy, rnd)
			if got == nil {
				t.Fatalf("%s returned nil", strategy)
			}
			if got.Code != "a" && got.Code != "b" && got.Code != "c" {
				t.Fatalf("%s picked unknown node %s", strategy, got.Code)
			}
		}
	}

	// weight：单节点权重为 0 时仍应返回（按 1 计）
	single := []model.GostNode{{Base: model.Base{Code: "z"}, Weight: 0}}
	if got := PickByStrategy(single, model.GOST_NODE_GROUP_STRATEGY_WEIGHT, rnd); got == nil || got.Code != "z" {
		t.Fatalf("weight single zero-weight node should pick z, got %v", got)
	}

	// 未知策略回退 priority
	if got := PickByStrategy(nodes, "unknown", rnd); got == nil || got.Code != "b" {
		t.Fatalf("unknown strategy should fall back to priority (b), got %v", got)
	}
}
