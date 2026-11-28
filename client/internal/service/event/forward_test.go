package event

import "testing"

func TestParsePortRange(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"", 0},
		{"8000-8010", 11},
		{"8000-8000", 1},
		{"8000", 0},
		{"8010-8000", 0},
		{"0-10", 0},
		{"1-2000", 0},
		{" 8000 - 8002 ", 3},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			got := parsePortRange(c.in)
			if len(got) != c.want {
				t.Fatalf("in=%q got %d ports want %d", c.in, len(got), c.want)
			}
		})
	}
	if ports := parsePortRange("8000-8002"); len(ports) != 3 || ports[0] != 8000 || ports[2] != 8002 {
		t.Fatalf("unexpected ports %v", ports)
	}
}
