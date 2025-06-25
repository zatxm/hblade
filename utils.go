package hblade

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
