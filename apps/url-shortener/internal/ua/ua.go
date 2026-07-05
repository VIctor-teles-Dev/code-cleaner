// Package ua classifica user-agents em (navegador, dispositivo) com heurísticas
// simples de string. Preciso o suficiente para a analytics do dashboard e sem
// dependência externa — coerente com o mínimo-de-deps do repo.
package ua

import "strings"

// Parse retorna o navegador e o dispositivo de um user-agent.
func Parse(userAgent string) (browser, device string) {
	return Browser(userAgent), Device(userAgent)
}

// Browser classifica o navegador. A ordem importa: Edge antes de Chrome (o UA
// do Edge contém "Chrome") e Safari por último (Chrome também contém "Safari").
func Browser(userAgent string) string {
	s := strings.ToLower(userAgent)
	switch {
	case s == "":
		return ""
	case containsAny(s, "bot", "crawler", "spider", "slurp"):
		return "Bot"
	case strings.Contains(s, "edg"):
		return "Edge"
	case strings.Contains(s, "opr") || strings.Contains(s, "opera"):
		return "Opera"
	case strings.Contains(s, "firefox"):
		return "Firefox"
	case strings.Contains(s, "chrome") || strings.Contains(s, "chromium"):
		return "Chrome"
	case strings.Contains(s, "safari"):
		return "Safari"
	default:
		return "Outro"
	}
}

// Device classifica o dispositivo. Tablets Android normalmente omitem "Mobile"
// no UA, então "android" sem "mobi" é tratado como Tablet.
func Device(userAgent string) string {
	s := strings.ToLower(userAgent)
	switch {
	case s == "":
		return ""
	case strings.Contains(s, "ipad") || strings.Contains(s, "tablet"):
		return "Tablet"
	case strings.Contains(s, "android") && !strings.Contains(s, "mobi"):
		return "Tablet"
	case containsAny(s, "mobi", "iphone", "android", "ipod"):
		return "Mobile"
	default:
		return "Desktop"
	}
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
