package util

import (
	"bufio"
	"log"
	"os"
)

var Logger = log.New(bufio.NewWriter(os.Stdout), "", log.Lshortfile)

func init() {
	Logger.SetOutput(os.Stdout)
}
