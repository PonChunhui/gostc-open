package dnsfollow

import "testing"

func TestZoneOf(t *testing.T) {
	cases := []struct{ domain, configured, want string }{
		{"a.b.example.com", "", "example.com"},
		{"example.com", "", "example.com"},
		{"a.example.com", "custom.org", "custom.org"},
		{"x.cn", "", "x.cn"},
	}
	for _, c := range cases {
		if got := ZoneOf(c.domain, c.configured); got != c.want {
			t.Fatalf("ZoneOf(%q,%q)=%q want %q", c.domain, c.configured, got, c.want)
		}
	}
}

func TestSubOf(t *testing.T) {
	cases := []struct{ domain, zone, want string }{
		{"a.example.com", "example.com", "a"},
		{"a.b.example.com", "example.com", "a.b"},
		{"example.com", "example.com", "@"},
	}
	for _, c := range cases {
		if got := SubOf(c.domain, c.zone); got != c.want {
			t.Fatalf("SubOf(%q,%q)=%q want %q", c.domain, c.zone, got, c.want)
		}
	}
}
