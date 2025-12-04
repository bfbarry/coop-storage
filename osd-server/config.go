package main

import (
	"os"
	"strconv"
)

var (
	PORT            string
	UPLOADDIR       string
	MAXUPLOADSIZE   int64
	METADATASERVERURL string
)

func init() {
	PORT = getEnv("PORT", "8280")
	UPLOADDIR = getEnv("UPLOAD_DIR", "./uploads")
	
	maxSizeStr := getEnv("MAX_UPLOAD_SIZE", "10485760")
	maxSize, err := strconv.ParseInt(maxSizeStr, 10, 64)
	if err != nil {
		maxSize = 10 << 20
	}
	MAXUPLOADSIZE = maxSize
	
	METADATASERVERURL = getEnv("METADATA_SERVER_URL", "http://localhost:8080")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

