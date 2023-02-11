package utils

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gen2brain/go-unarr"
)

// TODO don't forget to set this
const IN_DOCKER bool = false

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func RemoveFilesThatDontMatchPrefix(prefix string) {
	files, err := filepath.Glob("*")
	if err != nil {
		log.Printf("Error while getting files\n")
		return
	}
	for _, f := range files {
		if strings.HasPrefix(f, prefix) {
			continue
		}
		if err := os.RemoveAll(f); err != nil {
			log.Printf("Error while removing file %+v with err %+v\n", f, err)
		}
	}
}

//create a tmp .txt file, that it will populate with the tmpFile
//it will then call the exe, which is expected to use the tmpFile
//if all goes well, it returns the result written by the exe into the tmpFile

// this function doesn't log. The caller is responsible for logging
func ExecuteProcess(loggerRole int, processType FileType, tmpFile File, exeFile RunnableFile) ([]byte, error) {
	err := CreateAndWriteToFile(tmpFile.Name, tmpFile.Content)
	if err != nil {
		return nil, fmt.Errorf("error while creating the temporary file that contains the job contents for %+v process", processType)
	}

	_, err = exec.Command(exeFile.RunCmd).Output()
	if err != nil {
		return nil, fmt.Errorf("error while executing %+v process", processType)
	}

	//now need to read from this file the resulting data
	data, err := os.ReadFile(tmpFile.Name)
	if err != nil {
		return nil, fmt.Errorf("error while reading from the %+v process", processType)
	}
	return data, nil
}

func CreateAndWriteToFile(name string, data []byte) error {
	if err := os.MkdirAll("./"+filepath.Dir(name), os.ModePerm); err != nil {
		return err
	}

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

func UnzipSource(source, destination string) error {
	a, err := unarr.NewArchive(source)
	if err != nil {
		panic(err)
	}
	defer a.Close()

	_, err = a.Extract(destination)
	if err != nil {
		panic(err)
	}
	return nil
}

// func main() {
// 	err := unzipSource("testFolder.zip", "")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// }
