package model

import "encoding/json"

// parseStringList 解析 JSON 数组字符串；非法或空返回空切片。
func parseStringList(s string) []string {
	if s == "" {
		return []string{}
	}
	var list []string
	if err := json.Unmarshal([]byte(s), &list); err != nil {
		return []string{}
	}
	return list
}

// MarshalStringList 将字符串切片序列化为 JSON 数组字符串。
func MarshalStringList(list []string) string {
	if list == nil {
		list = []string{}
	}
	b, _ := json.Marshal(list)
	return string(b)
}
