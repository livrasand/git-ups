package config

import "testing"

func TestGatewayURL(t *testing.T) {
	cases := []struct {
		name    string
		env     string
		want    string
		wantErr bool
	}{
		{"unset uses default", "", DefaultGatewayURL, false},
		{"custom https", "https://gw.example.org", "https://gw.example.org", false},
		{"trailing slash removed", "https://gw.example.org/", "https://gw.example.org", false},
		{"custom http with port", "http://127.0.0.1:8080", "http://127.0.0.1:8080", false},
		{"missing scheme rejected", "gw.example.org", "", true},
		{"non-http scheme rejected", "ftp://gw.example.org", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("GITGOST_SERVER", tc.env)
			got, err := GatewayURL()
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
