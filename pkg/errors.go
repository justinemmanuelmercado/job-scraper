package errorHandler

import (
	"log"
	"strings"
)

func HandleErrorWithSection(err error, message string, section string) {
	if err != nil {
		log.Printf("[%s] - %s: %v\n", strings.ToUpper(section), message, err)
	}
}

func HandleError(err error, message string) {
	HandleErrorWithSection(err, message, "Default")
}
