package internal

import "fmt"

func GenerateHeaders(ssoCookie string) map[string]string {
	return map[string]string{
		"accept":             "*/*",
		"accept-language":    "en-US,en;q=0.9,es;q=0.8",
		"sec-ch-ua":          `"Not /ABrand";v="99", "Google Chrome";v="115", "Chromium";v="115"`,
		"sec-ch-ua-mobile":   "?0",
		"sec-ch-ua-platform": `"Windows"`,
		"sec-fetch-dest":     "empty",
		"sec-fetch-mode":     "cors",
		"sec-fetch-site":     "same-origin",
		"x-requested-with":   "XMLHttpRequest",
		"cookie":             fmt.Sprintf("ACT_SSO_COOKIE=%s", ssoCookie),
		"Referrer-Policy":    "strict-origin-when-cross-origin",
	}
}
