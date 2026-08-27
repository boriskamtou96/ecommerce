package utils

import "strings"

// CDNURL turns a storage key ("products/12/ab34.jpg") into the public URL
// served by the CDN. The key alone is what gets persisted, so the CDN
// host can change without touching stored data.
func CDNURL(baseURL, key string) string {
	if key == "" {
		return ""
	}

	// Already absolute (legacy rows, or an external image).
	if strings.HasPrefix(key, "http://") || strings.HasPrefix(key, "https://") {
		return key
	}

	return strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(key, "/")
}
