package config

import "strings"

func stripHTTPPrefix(addr string) string {
	if strings.HasPrefix(addr, "http://") {
		return addr[7:]
	}
	if strings.HasPrefix(addr, "https://") {
		return addr[8:]
	}
	return addr
}
