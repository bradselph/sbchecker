package services

import (
	"fmt"
)

func GenerateHeaders(ssoCookie string) map[string]string {
	return map[string]string{
		"accept":           "*/*",
		"sec-fetch-mode":   "cors",
		"x-requested-with": "XMLHttpRequest",
		"cookie":           fmt.Sprintf("ACT_SSO_COOKIE=%s", ssoCookie),
	}
}
