package logger

import (
	"log"
	"os"
)

var (
	Logger *log.Logger
)

const (
	path string = "swap.log"
)

func Init(logtofile bool) {
	if logtofile {
		f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening log file: %v", err)
		}
		Logger = log.New(f, "", log.LstdFlags|log.Lshortfile)
	} else {
		Logger = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
	}
}
