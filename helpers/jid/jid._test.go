package jid

import "testing"

func TestFromAddress(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"info@xn--gckvb8fzb.com", "info@マリウス.com"},
		{"alice@example.com", "alice@example.com"},
		{"info@マリウス.com", "info@マリウス.com"},
		{"user@xn--mnchen-3ya.de", "user@münchen.de"},
		{"user@xn--mlaga-xqa.es", "user@málaga.es"},
		{"no-at-sign", "no-at-sign"},
		{"local@", "local@"},
	}

	for _, c := range cases {
		if got := FromAddress(c.in); got != c.want {
			t.Errorf("FromAddress(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
