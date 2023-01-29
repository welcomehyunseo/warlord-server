package server

import (
	"fmt"
	"time"
)

const (
	InfoLevel  = "Info"
	DebugLevel = "Debug"
	ErrorLevel = "Error"
)

type LgElement struct {
	key   string
	value any
}

func NewLgElement(
	key string, value any,
) *LgElement {
	return &LgElement{
		key, value,
	}
}

func (le *LgElement) String() string {
	return fmt.Sprintf(
		"%s: %+v",
		le.key, le.value,
	)
}

func lgElementsToString(
	elements []*LgElement,
) string {
	var str string
	length := len(elements)
	for i, element := range elements {
		str += element.String()

		if i+1 < length {
			str += ", "
		}
	}
	return str
}

type Logger struct {
	prefix string
}

func NewLogger(
	elements ...*LgElement,
) *Logger {
	prefix := lgElementsToString(elements)
	return &Logger{
		prefix: prefix,
	}
}

func (c *Logger) Info(
	message string,
	elements ...*LgElement,
) {
	ms := time.Now().UnixMilli()
	str := lgElementsToString(elements)
	fmt.Printf(
		"{ timestamp: %d, level: %s, %s, message: %s, %s }\n",
		ms, InfoLevel, c.prefix, message, str,
	)
}

func (c *Logger) Debug(
	message string,
	elements ...*LgElement,
) {
	ms := time.Now().UnixMilli()
	str := lgElementsToString(elements)
	fmt.Printf(
		"{ timestamp: %d, level: %s, %s, message: %s, %s }\n",
		ms, DebugLevel, c.prefix, message, str,
	)
}

func (c *Logger) Error(
	err any,
) {
	ms := time.Now().UnixMilli()
	fmt.Printf(
		"{ timestamp: %d, level: %s, %s, error: %s }\n",
		ms, ErrorLevel, c.prefix, err,
	)
}
