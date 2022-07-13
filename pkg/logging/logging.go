package logging

import (
	"io"
	"log"
	"os"
)

// init
// spread logs between stdout and file
func Init() {
	err := os.MkdirAll("logs", 0755)

	if err != nil || os.IsExist(err) {
		panic("can't create log dir. no configured logging to files")
	} else {
		allFile, err := os.OpenFile("logs/all.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Fatalf("Failed to set-up logging system : %s", err.Error())
		}
		log.SetOutput(io.MultiWriter(allFile, os.Stdout))
	}

}
