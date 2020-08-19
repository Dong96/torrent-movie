package logger

import (
	"io"
	"log"
	"os"
)

var fname = "logger.log"
var mw io.Writer

//MultiLogWriter return writer for logging
func MultiLogWriter() io.Writer {
	return mw
}

func init() {
	file, err := os.OpenFile(fname, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}
	mw = io.MultiWriter(os.Stdout, file)

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(mw)
}
