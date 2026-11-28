// Package node_select 提供节点分组调度：按策略（优先级/权重/随机）从在线节点中挑选一个。
package node_select

import (
	"errors"
	"math/rand"
	"server/model"
	cache2 "server/repository/cache"
	"server/repository/query"
)

// PickByStrategy 纯函数：从候选节点按策略挑选一个；候选为空返回 nil。rnd 用于权重/随机策略。
func PickByStrategy(nodes []model.GostNode, strategy string, rnd *rand.Rand) *model.GostNode {
	if len(nodes) == 0 {
		return nil
	}
	switch strategy {
	case model.GOST_NODE_GROUP_STRATEGY_WEIGHT:
		total := 0
		for _, n := range nodes {
			w := n.Weight
			if w < 1 {
				w = 1
			}
			total += w
		}
		r := rnd.Intn(total)
		for i := range nodes {
			w := nodes[i].Weight
			if w < 1 {
				w = 1
			}
			if r < w {
				return &nodes[i]
			}
			r -= w
		}
		return &nodes[len(nodes)-1]
	case model.GOST_NODE_GROUP_STRATEGY_RANDOM:
		return &nodes[rnd.Intn(len(nodes))]
	default: // priority：越小越优先
		best := 0
		for i := range nodes {
			if nodes[i].Priority < nodes[best].Priority {
				best = i
			}
		}
		return &nodes[best]
	}
}

// SelectNode 从指定分组中挑选一个在线节点（排除 excludeCode）；无可用节点返回错误。
func SelectNode(tx *query.Query, groupCode, excludeCode string, rnd *rand.Rand) (*model.GostNode, error) {
	if groupCode == "" {
		return nil, errors.New("未指定节点分组")
	}
	group, _ := tx.GostNodeGroup.Where(tx.GostNodeGroup.Code.Eq(groupCode)).First()
	if group == nil {
		return nil, errors.New("节点分组不存在")
	}
	nodes, err := tx.GostNode.Where(tx.GostNode.GroupCode.Eq(groupCode)).Find()
	if err != nil {
		return nil, err
	}
	var candidates []model.GostNode
	for _, n := range nodes {
		if n.Code == excludeCode {
			continue
		}
		if !cache2.GetNodeOnline(n.Code) {
			continue
		}
		candidates = append(candidates, *n)
	}
	if len(candidates) == 0 {
		return nil, errors.New("分组内无在线可用节点")
	}
	return PickByStrategy(candidates, group.Strategy, rnd), nil
}
