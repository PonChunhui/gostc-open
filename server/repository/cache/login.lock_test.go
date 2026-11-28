package cache

import (
	"path/filepath"
	"server/global"
	"server/pkg/memory"
	"testing"
	"time"
)

func setupCache(t *testing.T) {
	t.Helper()
	global.Cache = memory.NewMemory(filepath.Join(t.TempDir(), "cache.db"), time.Hour)
}

func TestLoginLockout(t *testing.T) {
	setupCache(t)
	acct := "locktest"
	if LoginLocked(acct) {
		t.Fatal("should not be locked initially")
	}
	for i := 0; i < MaxLoginFails; i++ {
		LoginFail(acct)
	}
	if !LoginLocked(acct) {
		t.Fatalf("should be locked after %d failures", MaxLoginFails)
	}
	LoginSuccess(acct)
	if LoginLocked(acct) {
		t.Fatal("success should clear lock")
	}
}

func TestLoginFailBelowThreshold(t *testing.T) {
	setupCache(t)
	acct := "belowthreshold"
	for i := 0; i < MaxLoginFails-1; i++ {
		LoginFail(acct)
	}
	if LoginLocked(acct) {
		t.Fatal("should not lock below threshold")
	}
	LoginSuccess(acct)
	LoginFail(acct)
	if LoginLocked(acct) {
		t.Fatal("counter should reset after success")
	}
}
