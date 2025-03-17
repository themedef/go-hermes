package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type LogLevel string

const (
	DEBUG LogLevel = "DEBUG"
	INFO  LogLevel = "INFO"
	WARN  LogLevel = "WARN"
	ERROR LogLevel = "ERROR"
)

type Config struct {
	Enabled bool
	LogFile string
}

type Logger struct {
	config     Config
	fileLogger *log.Logger
	console    *log.Logger
	logFile    *os.File
	mutex      sync.Mutex
}

type LogMessage struct {
	Timestamp string   `json:"timestamp"`
	Level     LogLevel `json:"level"`
	Message   string   `json:"message"`
}

func NewLogger(config Config) (*Logger, error) {
	if !config.Enabled {
		return &Logger{config: config}, nil
	}

	var fileLogger *log.Logger
	var file *os.File
	var err error

	if config.LogFile != "" {
		file, err = os.OpenFile(config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("❌ Failed to open log file: %v", err)
		}
		fileLogger = log.New(file, "", log.LstdFlags)
	}

	consoleLogger := log.New(os.Stdout, "", log.LstdFlags)

	return &Logger{
		config:     config,
		fileLogger: fileLogger,
		console:    consoleLogger,
		logFile:    file,
	}, nil
}
func (l *Logger) log(level LogLevel, format string, v ...interface{}) {
	if !l.config.Enabled {
		return
	}

	message := fmt.Sprintf(format, v...)
	logEntry := LogMessage{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Level:     level,
		Message:   message,
	}

	logJSON, _ := json.Marshal(logEntry)

	l.mutex.Lock()
	defer l.mutex.Unlock()

	switch level {
	case DEBUG:
		l.console.Printf("\033[36m%s\033[0m", logJSON)
	case INFO:
		l.console.Printf("\033[32m%s\033[0m", logJSON)
	case WARN:
		l.console.Printf("\033[33m%s\033[0m", logJSON)
	case ERROR:
		l.console.Printf("\033[31m%s\033[0m", logJSON)
	}
	if l.fileLogger != nil {
		l.fileLogger.Println(string(logJSON))
	}
}

func (l *Logger) Debug(format string, v ...interface{}) {
	l.log(DEBUG, format, v...)
}

func (l *Logger) Info(format string, v ...interface{}) {
	l.log(INFO, format, v...)
}

func (l *Logger) Warn(format string, v ...interface{}) {
	l.log(WARN, format, v...)
}

func (l *Logger) Error(format string, v ...interface{}) {
	l.log(ERROR, format, v...)
}

func (l *Logger) Close() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}
