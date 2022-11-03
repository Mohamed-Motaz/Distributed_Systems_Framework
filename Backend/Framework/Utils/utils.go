package utils

import (
	"os"
)

const IN_DOCKER bool = false

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

