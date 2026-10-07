package execution

import (
	"net"
	"testing"
)

func TestPublicIP(t *testing.T) {
	for _, tc := range []struct {
		ip   string
		want bool
	}{
		{"127.0.0.1", false}, {"10.1.2.3", false}, {"172.17.0.2", false},
		{"169.254.169.254", false}, {"100.100.100.100", false},
		{"::1", false}, {"fc00::1", false}, {"2001:db8::1", false},
		{"8.8.8.8", true}, {"2606:4700:4700::1111", true},
	} {
		if got := publicIP(net.ParseIP(tc.ip)); got != tc.want {
			t.Errorf("%s: got %v, want %v", tc.ip, got, tc.want)
		}
	}
}
func TestHTTPURLValidation(t *testing.T) {
	for _, raw := range []string{"http://127.0.0.1/", "https://example.com/path?x=1"} {
		if err := validateHTTPURL(raw); err != nil {
			t.Errorf("valid URL %q: %v", raw, err)
		}
	}
	for _, raw := range []string{"file:///etc/passwd", "ftp://example.com", "example.com", "https://user:pass@example.com"} {
		if err := validateHTTPURL(raw); err == nil {
			t.Errorf("accepted %q", raw)
		}
	}
}
