package client_version

import "testing"

func TestInGray(t *testing.T) {
	cases := []struct {
		name    string
		percent int
		uid     string
		want    bool
	}{
		{"全量", 100, "u1", true},
		{"超百", 150, "u1", true},
		{"匿名全量", 10, "", true},
		{"零灰度", 0, "u1", false},
		{"负灰度", -5, "u1", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := inGray(c.percent, c.uid); got != c.want {
				t.Fatalf("inGray(%d,%q)=%v want %v", c.percent, c.uid, got, c.want)
			}
		})
	}
	// 灰度分桶稳定性：同一 uid 结果一致
	for i := 0; i < 5; i++ {
		if inGray(50, "stable-uid") != inGray(50, "stable-uid") {
			t.Fatal("灰度分桶不稳定")
		}
	}
	// 50% 灰度对大量 uid 应近似半数命中
	hit := 0
	for i := 0; i < 1000; i++ {
		if inGray(50, string(rune('a'+i%26))+string(rune('0'+i/26))) {
			hit++
		}
	}
	if hit < 300 || hit > 700 {
		t.Fatalf("50%%灰度命中数异常: %d", hit)
	}
}
