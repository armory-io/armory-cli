package util

import (
	"fmt"
	"os"
)

func TempAppFile(tmpDir, fileName, fileContent string) *os.File {
	tempFile, _ := os.CreateTemp(tmpDir, fileName)
	bytes, err := tempFile.Write([]byte(fileContent))
	if err != nil || bytes == 0 {
		fmt.Println("Could not write temp file.")
		return nil
	}
	return tempFile
}
