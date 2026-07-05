package ua

import "testing"

func TestParse(t *testing.T) {
	cases := []struct {
		name          string
		ua            string
		wantBrowser   string
		wantDevice    string
	}{
		{"chrome desktop", "Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 Chrome/120.0 Safari/537.36", "Chrome", "Desktop"},
		{"edge antes de chrome", "Mozilla/5.0 (Windows NT 10.0) Chrome/120 Safari/537 Edg/120.0", "Edge", "Desktop"},
		{"firefox", "Mozilla/5.0 (X11; Linux) Gecko/20100101 Firefox/121.0", "Firefox", "Desktop"},
		{"safari mac", "Mozilla/5.0 (Macintosh; Intel Mac OS X) Version/17.0 Safari/605.1", "Safari", "Desktop"},
		{"iphone", "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0) Version/17 Mobile/15E Safari/604.1", "Safari", "Mobile"},
		{"android phone", "Mozilla/5.0 (Linux; Android 13; Pixel 7) Chrome/120 Mobile Safari/537.36", "Chrome", "Mobile"},
		{"android tablet", "Mozilla/5.0 (Linux; Android 13; SM-X200) Chrome/120 Safari/537.36", "Chrome", "Tablet"},
		{"ipad", "Mozilla/5.0 (iPad; CPU OS 17_0) Version/17 Safari/604.1", "Safari", "Tablet"},
		{"googlebot", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)", "Bot", "Desktop"},
		{"vazio", "", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotBrowser, gotDevice := Parse(tc.ua)
			if gotBrowser != tc.wantBrowser {
				t.Errorf("browser = %q, want %q", gotBrowser, tc.wantBrowser)
			}
			if gotDevice != tc.wantDevice {
				t.Errorf("device = %q, want %q", gotDevice, tc.wantDevice)
			}
		})
	}
}
