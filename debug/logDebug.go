package debug

import (
	"log"
	"os"
)

func LogConfig(path string) {
	file, _ := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	log.SetOutput(file)
}
