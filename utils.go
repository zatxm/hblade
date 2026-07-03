package hblade

import "net/http"

type H map[string]any

func filterFlags(content string) string {
	for i := range content {
		char := content[i]
		if char == ' ' || char == ';' {
			return content[:i]
		}
	}
	return content
}

func hasRequestBody(method string) bool {
	return method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch
}
