package server

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go/types"
	"sync"
	"time"
)

type LogLevel int
type LoggerName string

var AlreadyLoggerExistError = errors.New("there are already existing logger")

const (
	DebugLevel    = LogLevel(2)
	InfoLevel     = LogLevel(1)
	ErrorLevel    = LogLevel(0)
	DebugLevelStr = "debug"
	InfoLevelStr  = "info"
	ErrorLevelStr = "error"
)

var globalLoggerConfigurator *LoggerConfigurator

func init() {
	var mutex sync.RWMutex
	globalLoggerConfigurator = &LoggerConfigurator{
		logLevel: DebugLevel,
		filter:   false,
		filters:  make(map[LoggerName]types.Nil),
		report:   false,
		mutex:    &mutex,
		loggers:  make(map[uuid.UUID]*Logger),
	}
}

type LoggerConfigurator struct {
	logLevel LogLevel
	filter   bool
	filters  map[LoggerName]types.Nil

	report bool

	mutex   *sync.RWMutex
	loggers map[uuid.UUID]*Logger
}

func NewLoggerConfigurator() *LoggerConfigurator {
	return globalLoggerConfigurator
}

func (lc *LoggerConfigurator) GetLogLevel() LogLevel {
	return lc.logLevel
}

func (lc *LoggerConfigurator) SetLogLevel(level LogLevel) {
	lc.logLevel = level
}

func (lc *LoggerConfigurator) IsFilter() bool {
	return lc.filter
}

func (lc *LoggerConfigurator) HasFilter(name LoggerName) bool {
	_, has := lc.filters[name]
	return has
}

func (lc *LoggerConfigurator) SetFilter(name LoggerName) {
	lc.filter = true
	lc.filters[name] = types.Nil{}
}

func (lc *LoggerConfigurator) IsReport() bool {
	return lc.report
}

func (lc *LoggerConfigurator) EnableReport() {
	lc.report = true
}

func (lc *LoggerConfigurator) DisableReport() {
	lc.report = false
}

func (lc *LoggerConfigurator) SetLogger(
	id uuid.UUID, logger *Logger,
) {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	//_, has := lc.loggers[id]
	//if has == true {
	//	panic(AlreadyLoggerExistError)
	//}
	lc.loggers[id] = logger

	report := lc.report
	if report == true {
		fmt.Println(lc)
	}

}

func (lc *LoggerConfigurator) Close(id uuid.UUID) {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	delete(lc.loggers, id)

	report := lc.report
	if report == true {
		fmt.Println(lc)
	}
}

func (lc *LoggerConfigurator) String() string {
	str := "[ "
	last := len(lc.loggers)
	count := 0
	for _, logger := range lc.loggers {
		str += logger.String()

		count++
		if count < last {
			str += ", "
		}
	}
	str += " ]"
	return fmt.Sprintf(
		"{ name: %s, loggers: %s }",
		"logger-configurator", str,
	)
}

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
	id     uuid.UUID
	name   LoggerName
	prefix string
}

func NewLogger(
	name LoggerName,
	elements ...*LgElement,
) *Logger {
	prefix := lgElementsToString(elements)
	id, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}
	logger := &Logger{
		id:     id,
		name:   name,
		prefix: prefix,
	}
	globalLoggerConfigurator.SetLogger(id, logger)
	return logger
}

func (c *Logger) Close() {
	id := c.id
	globalLoggerConfigurator.Close(id)
}

func (c *Logger) Debug(
	message string,
	elements ...*LgElement,
) {
	logLevel := globalLoggerConfigurator.GetLogLevel()
	if logLevel < DebugLevel {
		return
	}
	filter := globalLoggerConfigurator.IsFilter()
	name := c.name
	if filter == true && globalLoggerConfigurator.HasFilter(name) == false {
		return
	}

	prefix := c.prefix
	ms := time.Now().UnixMilli()
	str := lgElementsToString(elements)
	fmt.Printf(
		"{ timestamp: %d, level: %s, name: %s, %s, message: %s, %s }\n",
		ms, DebugLevelStr, name, prefix, message, str,
	)
}

func (c *Logger) Info(
	message string,
	elements ...*LgElement,
) {
	logLevel := globalLoggerConfigurator.GetLogLevel()
	if logLevel < InfoLevel {
		return
	}
	filter := globalLoggerConfigurator.IsFilter()
	name := c.name
	if filter == true && globalLoggerConfigurator.HasFilter(name) == false {
		return
	}

	prefix := c.prefix
	ms := time.Now().UnixMilli()
	str := lgElementsToString(elements)
	fmt.Printf(
		"{ timestamp: %d, level: %s, name: %s, %s, message: %s, %s }\n",
		ms, InfoLevelStr, name, prefix, message, str,
	)
}

func (c *Logger) Error(
	err any,
) {
	logLevel := globalLoggerConfigurator.GetLogLevel()
	if logLevel < ErrorLevel {
		return
	}
	filter := globalLoggerConfigurator.IsFilter()
	name := c.name
	if filter == true && globalLoggerConfigurator.HasFilter(name) == false {
		return
	}

	prefix := c.prefix
	ms := time.Now().UnixMilli()
	fmt.Printf(
		"{ timestamp: %d, level: %s, name: %s, %s, error: %s }\n",
		ms, ErrorLevelStr, name, prefix, err,
	)
}

func (c *Logger) String() string {
	return fmt.Sprintf(
		"{ id: %s, name: %s, prefix: %s }",
		c.id, c.name, c.prefix,
	)
}
