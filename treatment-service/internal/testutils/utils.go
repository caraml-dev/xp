package testutils

import (
	"bytes"
	"log"
)

func CaptureStderrLogs(f func()) string {
	var buf bytes.Buffer

	writer := log.Writer()
	log.SetOutput(&buf)
	f()
	log.SetOutput(writer)

	return buf.String()
}
