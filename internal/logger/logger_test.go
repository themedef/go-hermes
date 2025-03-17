package logger

import (
	"os"
	"testing"
)

func TestLoggerWithFile(t *testing.T) {
	logFile := "test.log"
	defer func() {
		if err := os.Remove(logFile); err != nil {
			t.Logf("Failed to remove log file: %v", err)
		}
	}()

	config := Config{
		Enabled: true,
		LogFile: logFile,
	}
	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	if !logger.config.Enabled {
		t.Errorf("Expected logger to be enabled")
	}

	logger.Info("Test message")

	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Errorf("Log file was not created")
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

	if logger.config.Enabled {
		t.Errorf("Expected logger to be disabled")
	}

	logger.Info("This should not be logged")
}

func TestLoggerFileNotCreatedWhenDisabled(t *testing.T) {
	logFile := "disabled.log"
	defer func() {
		if err := os.Remove(logFile); err != nil {
			t.Logf("Failed to remove log file: %v", err)
		}
	}()

	config := Config{
		Enabled: false,
		LogFile: logFile,
	}
	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	if _, err := os.Stat(logFile); !os.IsNotExist(err) {
		t.Errorf("Log file should not be created when logging is disabled")
	}

	if logger.config.Enabled {
		t.Errorf("Expected logger to be disabled")
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
