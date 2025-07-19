package fileio

import (
	"fmt"
	"os"
)

// CreateDirKeepContent creates a directory and does not remove any pre-existing contents
// It exits the application with an error code in case of failure
func CreateDirKeepContent(dir string) {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		panic(fmt.Sprintf("could not create directory %s", err))
	}
}

// CreateEmptyDir creates a directory after clearing it if it existed already
// It exits the application with an error code in case of failure
func CreateEmptyDir(dir string) {
	err := os.RemoveAll(dir)
	if err != nil {
		panic(fmt.Sprintf("could not empty directory %s", err))
	}
	CreateDirKeepContent(dir)
}
