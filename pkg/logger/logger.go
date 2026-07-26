package logger

import (
	"log"
	"os"
)

var logger *log.Logger

func init() {
	logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
}

func Info(msg string) {
	logger.Println("INFO: " + msg)
}

func Error(msg string) {
	logger.Println("ERROR: " + msg)
}
