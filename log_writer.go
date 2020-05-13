package main

import (
	"fmt"
	"log"
)

type LogWriter struct {
	Prefix string
	currentLine []byte
}

func (lw *LogWriter) SetDockerJobName(name string){
	lw.Prefix = fmt.Sprintf("DOCKER %s | ", name)
}

func (lw *LogWriter) Write(p []byte) (n int, err error) {
	for _, b := range p {
		if b == '\n' {
			// fmt.Printf("Writing log: %s\n", string(lw.currentLine))
			log.Printf("%s%s", lw.Prefix, string(lw.currentLine))
			lw.currentLine = make([]byte, 0)
		} else {
			lw.currentLine = append(lw.currentLine, b)
		}
	}
	return len(p), nil
}