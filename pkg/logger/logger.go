package logger

import (
	"fmt"
	"io"
	"log"
	"os"
)

var (
	file *os.File

	Output *log.Logger
	Info   *log.Logger
	Warn   *log.Logger
	Error  *log.Logger
)

func Init(outputFile string) error {
	if outputFile != "" {
		file, err := os.OpenFile(outputFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("error opening file: %v", err)
		}
		writer := io.MultiWriter(os.Stdout, file)
		Output = log.New(writer, "", 0)
	} else {
		Output = log.New(os.Stdout, "", 0)
	}

	Info = log.New(os.Stderr, "[INFO] ", 0)
	Warn = log.New(os.Stderr, "[WARN] ", 0)
	Error = log.New(os.Stderr, "[ERROR] ", 0)

	return nil
}

func Close() {
	if file != nil {
		file.Close()
	}
}
