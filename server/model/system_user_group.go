package model

import "strings"

// SystemUserGroup 用户组：按组细粒度控制可用节点/套餐、隧道数量上限、每月流量上限。
// 白名单采用逗号分隔字符串存储（镜像 GostNode.Rules/Tags），空表示不限制。
type SystemUserGroup struct {
	Base
	IndexValue   int    `gorm:"column:index_value;index;default:1000;comment:排序，升序"`
	Name         string `gorm:"column:name;index;comment:用户组名称"`
	Remark       string `gorm:"column:remark;comment:备注"`
	NodeCodes    string `gorm:"column:node_codes;type:text;comment:允许节点编号,逗号分隔,空=全部"`
	ConfigCodes  string `gorm:"column:config_codes;type:text;comment:允许套餐编号,逗号分隔,空=全部"`
	TunnelLimit  int    `gorm:"column:tunnel_limit;default:0;comment:隧道数量上限,0=无限制"`
	MonthTraffic int    `gorm:"column:month_traffic;default:0;comment:每月流量上限GB,0=无限制"`
}

// GetNodeCodes 返回允许使用的节点编号列表；空切片(nil)表示不限制(全部节点)。
func (g SystemUserGroup) GetNodeCodes() (result []string) {
	for _, code := range strings.Split(g.NodeCodes, ",") {
		if code == "" {
			continue
		}
		result = append(result, code)
	}
	return result
}

// GetConfigCodes 返回允许使用的套餐编号列表；空切片(nil)表示不限制(全部套餐)。
func (g SystemUserGroup) GetConfigCodes() (result []string) {
	for _, code := range strings.Split(g.ConfigCodes, ",") {
		if code == "" {
			continue
		}
		result = append(result, code)
	}
	return result
}
