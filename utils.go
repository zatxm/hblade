package hblade

import (
	"fmt"
	"os"
)

func filterFlags(content string) string {
	for i, char := range content {
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
			fmt.Printf("Environment variable PORT=\"%s\"", port)
			return ":" + port
		}
		fmt.Println("Environment variable PORT is undefined. Using port :7771 by default")
		return ":7771"
	case 1:
		return addr[0]
	default:
		panic("too many parameters")
	}
}
