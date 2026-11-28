package tunnel_adv

import "testing"

func TestValidateHealthCheck(t *testing.T) {
	cases := []struct {
		name      string
		hcType    string
		path      string
		interval  int
		timeout   int
		maxFailed int
		wantErr   bool
	}{
		{"空类型不启用", "", "", 0, 0, 0, false},
		{"tcp合法", "tcp", "", 10, 3, 3, false},
		{"http带路径", "http", "/health", 10, 3, 3, false},
		{"http缺路径", "http", "", 10, 3, 3, true},
		{"非法类型", "grpc", "", 10, 3, 3, true},
		{"间隔非正", "tcp", "", 0, 3, 3, true},
		{"超时为负", "tcp", "", 10, -1, 3, true},
		{"失败次数为负", "tcp", "", 10, 3, -1, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := ValidateHealthCheck(c.hcType, c.path, c.interval, c.timeout, c.maxFailed)
			if (err != nil) != c.wantErr {
				t.Fatalf("err=%v wantErr=%v", err, c.wantErr)
			}
		})
	}
}

func TestValidatePortRange(t *testing.T) {
	cases := []struct {
		in      string
		wantErr bool
	}{
		{"", false},
		{"8000-8010", false},
		{"8000-8000", false},
		{"8000", true},
		{"8010-8000", true},
		{"0-10", true},
		{"a-b", true},
		{"1-2000", true},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			err := ValidatePortRange(c.in)
			if (err != nil) != c.wantErr {
				t.Fatalf("in=%q err=%v wantErr=%v", c.in, err, c.wantErr)
			}
		})
	}
}

func TestValidateMultiplexer(t *testing.T) {
	for _, m := range []string{"", "httpconnect", "tcpmux"} {
		if err := ValidateMultiplexer(m); err != nil {
			t.Fatalf("valid %q got err %v", m, err)
		}
	}
	if err := ValidateMultiplexer("udp"); err == nil {
		t.Fatal("expected error for udp")
	}
}

func TestValidateBandwidth(t *testing.T) {
	cases := []struct {
		limit   string
		mode    string
		wantErr bool
	}{
		{"", "", false},
		{"1MB", "client", false},
		{"512KB", "server", false},
		{"2GB", "", false},
		{"1MB", "both", true},
		{"1TB", "", true},
		{"0MB", "", true},
		{"abc", "", true},
	}
	for _, c := range cases {
		t.Run(c.limit+"/"+c.mode, func(t *testing.T) {
			err := ValidateBandwidth(c.limit, c.mode)
			if (err != nil) != c.wantErr {
				t.Fatalf("limit=%q mode=%q err=%v wantErr=%v", c.limit, c.mode, err, c.wantErr)
			}
		})
	}
}
