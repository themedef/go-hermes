package logger

import (
	"os"
	"testing"
)

func TestLoggerWithFile(t *testing.T) {
	logFile := "test.log"
	defer func() {
		_ = os.Remove(logFile)
	}()

	config := Config{
		Enabled:    true,
		LogFile:    logFile,
		BufferSize: 100,
		MinLevel:   INFO,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.Info("Test message")
	err = logger.Close()
	if err != nil {
		t.Fatalf("Failed to close logger: %v", err)
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	if len(data) == 0 {
		t.Errorf("Log file is empty")
	}
}

func TestLoggerWithoutFile(t *testing.T) {
	config := Config{
		Enabled: false,
		LogFile: "",
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.Info("This should not be logged")

	err = logger.Close()
	if err != nil {
		t.Fatalf("Expected no error on close: %v", err)
	}
}
func TestLoggerFileNotCreatedWhenDisabled(t *testing.T) {
	logFile := "disabled.log"
	defer func() {
		_ = os.Remove(logFile)
	}()

	config := Config{
		Enabled: false,
		LogFile: logFile,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	_ = logger.Close()

	if _, err := os.Stat(logFile); !os.IsNotExist(err) {
		t.Errorf("Log file should not be created when logging is disabled")
	}
}

func TestLoggerWithInvalidFile(t *testing.T) {
	config := Config{
		Enabled: true,
		LogFile: "/invalid/path/test.log",
	}

	_, err := NewLogger(config)
	if err == nil {
		t.Errorf("Expected error when creating logger with invalid file path")
	}
}
