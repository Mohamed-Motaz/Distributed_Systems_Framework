package utils

import (
	"log"
	"os"
	"strings"
)

const IN_DOCKER bool = true

//THIS METHOD SHOULD BE CALLED ONLY ONCE
func InitDirectories(dataFolder string) {

	err := os.MkdirAll(dataFolder, os.ModePerm)
	if err != nil {
		log.Fatalf("Unable to initialize directories %v\n", err)
	}

	log.Printf("Done initializing directories\n\n")
}

// GetStringInBetween Returns empty string if no start string found
func GetStringInBetween(str string, start string, end string) (result string) {
	s := strings.Index(str, start)
	if s == -1 {
		return result
	}
	newS := str[s+len(start):]
	e := strings.Index(newS, end)
	if e == -1 {
		return result
	}
	result = newS[:e]
	return result
}

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
