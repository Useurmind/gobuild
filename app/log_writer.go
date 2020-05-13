package main

import (
	"log"
)

type LogWriter struct {
	JobName string
	currentLine []byte
}

func (lw LogWriter) Write(p []byte) (n int, err error) {
	for _, b := range p {
		if b == '\n' {
			log.Printf("DOCKER %s | %s", lw.JobName, string(lw.currentLine))
			lw.currentLine = make([]byte, 0)
		} else {
			lw.currentLine = append(lw.currentLine, b)
		}
	}
	return len(p), nil
}