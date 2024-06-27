package services

import (
	"fmt"
)

func GenerateHeaders(ssoCookie string) map[string]string {
	return map[string]string{
		"accept":         "*/*",
		"sec-fetch-mode": "cors",
		"cookie":         fmt.Sprintf("ACT_SSO_COOKIE=%s", ssoCookie),
	}
}

func GeneratePostHeaders(ssoCookie string) map[string]string {
	headers := GenerateHeaders(ssoCookie)
	headers["Content-Type"] = "application/x-www-form-urlencoded"
	headers["x-requested-with"] = "XMLHttpRequest"
	return headers
}
