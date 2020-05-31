package sys

import (
	"fmt"
	"io"
	"log"
	"os"
)

var (
	Log   *log.Logger
	Debug *log.Logger
)

func init() {
	name := fmt.Sprintf("%s.%s", SelfPath(), "log")
	file, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 666)
	if err != nil {
		log.Fatalln("Failed to open error log file:", err)
	}

	Debug = log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	Log = log.New(io.MultiWriter(file, os.Stderr), "LOG: ", log.Ldate|log.Ltime|log.Lshortfile)
}
