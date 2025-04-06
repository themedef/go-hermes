package logger

import (
	"encoding/json"
	"fmt"
	"github.com/themedef/go-hermes/internal/contracts"
	"log"
	"os"
	"strings"
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

var levelPriority = map[LogLevel]int{
	DEBUG: 1,
	INFO:  2,
	WARN:  3,
	ERROR: 4,
}

type Config struct {
	Enabled    bool
	LogFile    string
	BufferSize int
	MinLevel   LogLevel
}

type Logger struct {
	config     Config
	fileLogger *log.Logger
	console    *log.Logger
	logFile    *os.File
	logCh      chan LogMessage
	wg         sync.WaitGroup
}

type LogMessage struct {
	Timestamp string   `json:"timestamp"`
	Level     LogLevel `json:"level"`
	Message   string   `json:"message"`
}

func NewLogger(config Config) (contracts.LoggerHandler, error) {
	if !config.Enabled {
		return &Logger{config: config}, nil
	}

	if config.BufferSize <= 0 {
		config.BufferSize = 1000
	}
	if _, ok := levelPriority[config.MinLevel]; !ok {
		config.MinLevel = DEBUG
	}

	var fileLogger *log.Logger
	var file *os.File
	var err error

	if config.LogFile != "" {
		file, err = os.OpenFile(config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %v", err)
		}
		fileLogger = log.New(file, "", 0)
	}

	consoleLogger := log.New(os.Stdout, "", 0)

	logger := &Logger{
		config:     config,
		fileLogger: fileLogger,
		console:    consoleLogger,
		logFile:    file,
		logCh:      make(chan LogMessage, config.BufferSize),
	}

	logger.wg.Add(1)
	go logger.worker()

	return logger, nil
}

func (l *Logger) shouldLog(level LogLevel) bool {
	return levelPriority[level] >= levelPriority[l.config.MinLevel]
}

func (l *Logger) log(level LogLevel, msg string) {
	if !l.config.Enabled || !l.shouldLog(level) {
		return
	}

	entry := LogMessage{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Level:     level,
		Message:   msg,
	}

	select {
	case l.logCh <- entry:
	default:
	}
}

func (l *Logger) Debug(args ...interface{}) {
	l.log(DEBUG, strings.TrimSpace(fmt.Sprintln(args...)))
}

func (l *Logger) Info(args ...interface{}) {
	l.log(INFO, strings.TrimSpace(fmt.Sprintln(args...)))
}

func (l *Logger) Warn(args ...interface{}) {
	l.log(WARN, strings.TrimSpace(fmt.Sprintln(args...)))
}

func (l *Logger) Error(args ...interface{}) {
	l.log(ERROR, strings.TrimSpace(fmt.Sprintln(args...)))
}

func (l *Logger) worker() {
	defer l.wg.Done()
	for entry := range l.logCh {
		data, _ := json.Marshal(entry)

		switch entry.Level {
		case DEBUG:
			l.console.Printf("\033[36m%s\033[0m\n", data)
		case INFO:
			l.console.Printf("\033[32m%s\033[0m\n", data)
		case WARN:
			l.console.Printf("\033[33m%s\033[0m\n", data)
		case ERROR:
			l.console.Printf("\033[31m%s\033[0m\n", data)
		}

		if l.fileLogger != nil {
			l.fileLogger.Println(string(data))
		}
	}
}

func (l *Logger) Close() error {
	if !l.config.Enabled {
		return nil
	}

	close(l.logCh)
	l.wg.Wait()

	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}
