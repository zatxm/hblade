package hblade

import (
	"fmt"
	"os"

	"github.com/zeebo/xxh3"
)

func filterFlags(content string) string {
	for i, char := range content {
		if char == ' ' || char == ';' {
			return content[:i]
		}
	}
	return content
}

// Bytes hashes the given byte slice.
func BytesHash(in []byte) uint64 {
	return xxh3.Hash(in)
}

// String hashes the given string.
func StringHash(in string) uint64 {
	return xxh3.HashString(in)
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
