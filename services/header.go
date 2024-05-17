package services

import (
	"fmt"
)

// GenerateHeaders generates the headers for the request to the Activision API.
func GenerateHeaders(ssoCookie string) map[string]string {
	return map[string]string{
		"accept":           "*/*",
		"sec-fetch-mode":   "cors",
		"x-requested-with": "XMLHttpRequest",
		"cookie":           fmt.Sprintf("ACT_SSO_COOKIE=%s", ssoCookie),
	}
}
