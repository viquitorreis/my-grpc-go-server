package main

import (
	"fmt"
	"time"
)

type logWriter struct{}

func (l *logWriter) Write(p []byte) (n int, err error) {
	return fmt.Print(time.Now().Format("2006-01-02 15:04:05") + " " + string(p))
}
