package service

import (
	"server/model"
	"testing"
	"time"
)

func TestApplyGroupGrant(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("新升级限时组_记录原组", func(t *testing.T) {
		user := &model.SystemUser{GroupCode: "free"}
		applyGroupGrant(user, "pro", 30, now)
		if user.GroupCode != "pro" || user.PrevGroupCode != "free" {
			t.Fatalf("got group=%s prev=%s", user.GroupCode, user.PrevGroupCode)
		}
		if user.GroupExpAt != now.Unix()+30*86400 {
			t.Fatalf("got exp=%d", user.GroupExpAt)
		}
	})

	t.Run("无组用户升级_回落为空", func(t *testing.T) {
		user := &model.SystemUser{}
		applyGroupGrant(user, "pro", 30, now)
		if user.GroupCode != "pro" || user.PrevGroupCode != "" {
			t.Fatalf("got group=%s prev=%s", user.GroupCode, user.PrevGroupCode)
		}
	})

	t.Run("同组未到期_续期", func(t *testing.T) {
		exp := now.Unix() + 10*86400
		user := &model.SystemUser{GroupCode: "pro", GroupExpAt: exp, PrevGroupCode: "free"}
		applyGroupGrant(user, "pro", 30, now)
		if user.GroupExpAt != exp+30*86400 {
			t.Fatalf("got exp=%d want=%d", user.GroupExpAt, exp+30*86400)
		}
		if user.PrevGroupCode != "free" {
			t.Fatalf("prev changed: %s", user.PrevGroupCode)
		}
	})

	t.Run("永久卡_覆盖为永久", func(t *testing.T) {
		user := &model.SystemUser{GroupCode: "free", GroupExpAt: now.Unix() + 86400, PrevGroupCode: "base"}
		applyGroupGrant(user, "pro", 0, now)
		if user.GroupCode != "pro" || user.GroupExpAt != 0 || user.PrevGroupCode != "" {
			t.Fatalf("got group=%s exp=%d prev=%s", user.GroupCode, user.GroupExpAt, user.PrevGroupCode)
		}
	})

	t.Run("已是目标组且永久_保持不变", func(t *testing.T) {
		user := &model.SystemUser{GroupCode: "pro", GroupExpAt: 0, PrevGroupCode: "free"}
		applyGroupGrant(user, "pro", 30, now)
		if user.GroupCode != "pro" || user.GroupExpAt != 0 || user.PrevGroupCode != "free" {
			t.Fatalf("got group=%s exp=%d prev=%s", user.GroupCode, user.GroupExpAt, user.PrevGroupCode)
		}
	})

	t.Run("同组已过期_重新发放", func(t *testing.T) {
		user := &model.SystemUser{GroupCode: "pro", GroupExpAt: now.Unix() - 86400, PrevGroupCode: "free"}
		applyGroupGrant(user, "pro", 30, now)
		if user.GroupExpAt != now.Unix()+30*86400 || user.PrevGroupCode != "" {
			t.Fatalf("got exp=%d prev=%s", user.GroupExpAt, user.PrevGroupCode)
		}
	})
}
