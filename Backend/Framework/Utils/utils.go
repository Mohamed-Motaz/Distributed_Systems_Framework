package utils

import (
	"log"
	"os"
	"path/filepath"
)

const IN_DOCKER bool = false

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func RemoveFilesWithStartName(name string) {
	files, err := filepath.Glob(name + "*")
	if err != nil {
		log.Printf("Error while getting files that start with name %+v \n", name)
		return
	}
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			log.Printf("Error while removing file %+v \n", f)

		}
	}
}

func CreateAndWriteToFile(name string, data []byte) error {
	f, err := os.Create(name)
	if err != nil {
		return err
	}

	_, err = f.Write(data)
	if err != nil {
		return err
	}

	return f.Close()
}
