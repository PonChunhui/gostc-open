package tunnel_quota

import (
	"server/model"
	"server/repository/cache"
	"server/repository/query"
	"time"
)

// Refresh 计算隧道当前计费窗口内的历史流量(不含今日)并写入缓存。
// 循环计费窗口 = [ExpAt-Cycle天, ExpAt]，续费后窗口自然重置；
// 一次性/免费窗口 = [CreatedAt, 今天]，累计总量永不重置。
func Refresh(db *query.Query, code string, chargingType, cycle int, expAt int64, createdAt time.Time) {
	var start time.Time
	if chargingType == model.GOST_CONFIG_CHARGING_CUCLE_DAY && cycle > 0 {
		start = time.Unix(expAt, 0).AddDate(0, 0, -cycle)
	} else {
		start = createdAt
	}
	var obs cache.TunnelObs
	now := time.Now()
	if !start.After(now) {
		startDate := start.Format(time.DateOnly)
		yesterday := now.AddDate(0, 0, -1).Format(time.DateOnly)
		rows, _ := db.GostObs.Where(
			db.GostObs.OriginKind.Eq(model.GOST_OBS_ORIGIN_KIND_TUNNEL),
			db.GostObs.OriginCode.Eq(code),
			db.GostObs.EventDate.Gte(startDate),
			db.GostObs.EventDate.Lte(yesterday),
		).Find()
		for _, row := range rows {
			obs.InputBytes += row.InputBytes
			obs.OutputBytes += row.OutputBytes
		}
	}
	cache.SetTunnelObsLimit(code, obs)
}

// UsedBytes 返回隧道当前计费窗口已用流量(历史缓存 + 今日实时)。
func UsedBytes(code string) int64 {
	limit := cache.GetTunnelObsLimit(code)
	today := cache.GetTunnelObs(time.Now().Format(time.DateOnly), code)
	return limit.InputBytes + limit.OutputBytes + today.InputBytes + today.OutputBytes
}

// Exhausted 判断流量包是否已用完(limitTotal 单位 GB，0 = 无限制)。
func Exhausted(code string, limitTotal int) bool {
	if limitTotal <= 0 {
		return false
	}
	return UsedBytes(code) >= int64(limitTotal)*1024*1024*1024
}
