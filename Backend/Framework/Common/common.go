package common

import (
	logger "Framework/Logger"
	utils "Framework/Utils"

	"fmt"
	"os"
	"os/exec"
)

//create a tmp .txt file, that it will populate with the tmpFile
//it will then call the exe, which is expected to use the tmpFile
//if all goes well, it returns the result written by the exe into the tmpFile
func ExecuteProcess(loggerRole int, processType utils.FileType, tmpFile utils.File, exeFile utils.File) ([]byte, error) {
	err := utils.CreateAndWriteToFile(tmpFile.Name, tmpFile.Content)
	if err != nil {
		logger.LogError(loggerRole, logger.ESSENTIAL, "error while creating the temporary file that contains the job contents for %+v process locally %v", processType, err)
		return nil, fmt.Errorf("error while creating the temporary file that contains the job contents for %+v process", processType)
	}

	_, err = exec.Command("./" + exeFile.Name).Output()
	if err != nil {
		logger.LogError(loggerRole, logger.ESSENTIAL, "error while executing %+v process %+v", processType, err)
		return nil, fmt.Errorf("error while executing %+v process", processType)
	}

	//now need to read from this file the resulting data
	data, err := os.ReadFile(tmpFile.Name)
	if err != nil {
		logger.LogError(loggerRole, logger.ESSENTIAL, "error while reading from the %+v process %+v", processType, err)
		return nil, fmt.Errorf("error while reading from the %+v process", processType)
	}
	return data, nil
}
