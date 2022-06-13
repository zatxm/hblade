package hblade

import (
	"os"

	"go.uber.org/zap"
)

type H map[string]interface{}

func filterFlags(content string) string {
	for i := range content {
		char := content[i]
		if char == ' ' || char == ';' {
			return content[:i]
		}
	}
	return content
}

func resolveAddress(addr []string) string {
	switch len(addr) {
	case 0:
		if port := os.Getenv("PORT"); port != "" {
			Log.Info("Environment variable PORT",
				zap.String("port", port))
			return ":" + port
		}
		Log.Info("Environment variable PORT is undefined. Using port :7771 by default")
		return ":7771"
	case 1:
		return addr[0]
	default:
		panic("too many parameters")
	}
}
