package logger

import (
	"fmt"
	"log"
	"os"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

var (
	currentLevel = INFO
	logger      = log.New(os.Stdout, "", log.LstdFlags)
)

func SetLevel(level Level) {
	currentLevel = level
}

func Debug(format string, v ...interface{}) {
	if currentLevel <= DEBUG {
		logger.Printf("[DEBUG] "+format, v...)
	}
}

func Info(format string, v ...interface{}) {
	if currentLevel <= INFO {
		logger.Printf("[INFO] "+format, v...)
	}
}

func Warn(format string, v ...interface{}) {
	if currentLevel <= WARN {
		logger.Printf("[WARN] "+format, v...)
	}
}

func Error(format string, v ...interface{}) {
	if currentLevel <= ERROR {
		logger.Printf("[ERROR] "+format, v...)
	}
}

func Errorf(format string, v ...interface{}) {
	if currentLevel <= ERROR {
		logger.Printf("[ERROR] "+format, v...)
	}
}

func Fatal(format string, v ...interface{}) {
	logger.Fatalf("[FATAL] "+format, v...)
}

func ParseLevel(level string) (Level, error) {
	switch level {
	case "debug":
		return DEBUG, nil
	case "info":
		return INFO, nil
	case "warn":
		return WARN, nil
	case "error":
		return ERROR, nil
	default:
		return INFO, fmt.Errorf("unknown log level: %s", level)
	}
}
