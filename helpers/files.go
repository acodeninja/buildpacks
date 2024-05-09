package helpers

import (
	"io"
	"log"
	"os"
)

func CopyFile(source, destination string) error {
	sourceFile, err := os.Open(source)
	if err != nil {
		return err
	}

	defer func(sourceFile *os.File) {
		err := sourceFile.Close()
		if err != nil {
			log.Panic(err)
		}
	}(sourceFile)

	destinationFile, err := os.Create(destination)
	if err != nil {
		return err
	}

	defer func(destinationFile *os.File) {
		err := destinationFile.Close()
		if err != nil {
			log.Panic(err)
		}
	}(destinationFile)

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}

	return nil
}
