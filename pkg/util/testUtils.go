package util

import (
	"fmt"
	"io/ioutil"
	"os"
)

func TempAppFile(tmpDir, fileName, fileContent string) *os.File {
	tempFile, _ := ioutil.TempFile(tmpDir, fileName)
	bytes, err := tempFile.Write([]byte(fileContent))
	if err != nil || bytes == 0 {
		fmt.Println("Could not write temp file.")
		return nil
	}
	return tempFile
}
