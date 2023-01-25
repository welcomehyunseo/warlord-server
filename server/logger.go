package server

import (
	"fmt"
	"log"
)

type Logger struct {
	head string
}

func NewLogger(
	format string, v ...any,
) *Logger {
	return &Logger{
		head: fmt.Sprintf(format, v...),
	}
}

func (c *Logger) Info(msg string) {
	log.Printf("[INFO] %s { %s }\n", msg, c.head)
}

func (c *Logger) InfoWithVars(msg, format string, v ...any) {
	log.Printf("[INFO] %s { %s, %s }\n", msg, c.head, fmt.Sprintf(format, v...))
}

func (c *Logger) Error(err error) {
	log.Printf("[ERROR] %s, %s\n", c.head, err.Error())
}
