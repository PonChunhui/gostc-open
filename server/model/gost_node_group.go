package model

const (
	GOST_NODE_GROUP_STRATEGY_PRIORITY = "priority" // 按优先级（越小越优先）
	GOST_NODE_GROUP_STRATEGY_WEIGHT   = "weight"   // 按权重随机
	GOST_NODE_GROUP_STRATEGY_RANDOM   = "random"   // 均匀随机
)

// GostNodeGroup 节点分组：用于故障转移备节点选择与创建隧道时的自动调度。
type GostNodeGroup struct {
	Base
	Name     string `gorm:"column:name;index;comment:分组名称"`
	Strategy string `gorm:"column:strategy;default:'priority';comment:调度策略priority/weight/random"`
	Remark   string `gorm:"column:remark;comment:备注"`
}
