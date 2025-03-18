package logging

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	apexlog "github.com/apex/log"
	"github.com/apex/log/handlers/json"
	"github.com/apex/log/handlers/multi"
	"github.com/apex/log/handlers/text"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	// DefaultLogger is the singleton instance of our logger
	DefaultLogger *apexlog.Logger
	// LogFile is the path to the current log file
	LogFile string
	// AppDirName is the name of the application directory
	AppDirName string
)

// Fields is a type alias for log.Fields to make it easier to use
type Fields = apexlog.Fields

// SetAppName sets the application name for logging
func SetAppName(name string) {
	AppDirName = name
}

// getLogDirectory returns the appropriate log directory for the current OS
func getLogDirectory() string {
	switch runtime.GOOS {
	case "darwin":
		// macOS: ~/Library/Logs/SquareGolf Connector
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, "Library", "Logs", AppDirName)
		}
	case "windows":
		// Windows: %LOCALAPPDATA%\SquareGolf Connector\Logs
		appData := os.Getenv("LOCALAPPDATA")
		if appData != "" {
			return filepath.Join(appData, AppDirName, "Logs")
		}
	case "linux":
		// Linux: /var/log/squaregolf-connector (if root) or ~/.local/share/squaregolf-connector/logs
		if os.Getuid() == 0 {
			return filepath.Join("/var/log", AppDirName)
		}
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, ".local", "share", AppDirName, "logs")
		}
	}
	// Fallback to local logs directory
	return "logs"
}

// GetLogDirectory returns the path to the log directory
func GetLogDirectory() string {
	return getLogDirectory()
}

// Init initializes the logger with both console and file output
func Init() error {
	// Get the appropriate log directory
	logsDir := getLogDirectory()

	// Create logs directory if it doesn't exist
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return err
	}

	// Set up log file with rotation
	LogFile = filepath.Join(logsDir, "connector.log")
	rotator := &lumberjack.Logger{
		Filename:   LogFile,
		MaxSize:    5,    // megabytes
		MaxBackups: 5,    // number of backups to keep
		MaxAge:     28,   // days
		Compress:   true, // compress old files
	}

	// Create handlers
	consoleHandler := text.New(os.Stdout)
	fileHandler := json.New(rotator)

	// Create multi handler to write to both console and file
	handler := multi.New(
		consoleHandler,
		fileHandler,
	)

	// Create logger with our handler
	DefaultLogger = &apexlog.Logger{
		Handler: handler,
		Level:   apexlog.InfoLevel,
	}

	// Set up standard logger to use our implementation
	stdLogger := &stdLogger{apexLogger: DefaultLogger}

	// Configure the standard log package to use our logger
	log.SetOutput(stdLogger)
	log.SetFlags(0) // Disable standard log flags since we handle formatting ourselves

	return nil
}

// stdLogger implements the io.Writer interface to handle standard logging
type stdLogger struct {
	apexLogger *apexlog.Logger
}

// Write implements the io.Writer interface
func (l *stdLogger) Write(p []byte) (n int, err error) {
	// Get the caller's file and line number
	_, file, line, ok := runtime.Caller(3) // Skip 3 frames to get to the actual caller
	if !ok {
		file = "unknown"
		line = 0
	}

	// Create a new entry with the caller's information
	entry := l.apexLogger.WithFields(apexlog.Fields{
		"file": filepath.Base(file),
		"line": line,
	})

	// Log the message
	entry.Info(string(p))

	return len(p), nil
}

// Info logs an info message
func Info(msg string) {
	DefaultLogger.Info(msg)
}

// Error logs an error message
func Error(msg string) {
	DefaultLogger.Error(msg)
}

// Debug logs a debug message
func Debug(msg string) {
	DefaultLogger.Debug(msg)
}

// Warn logs a warning message
func Warn(msg string) {
	DefaultLogger.Warn(msg)
}

// WithField adds a field to the log entry
func WithField(key string, value interface{}) *apexlog.Entry {
	return DefaultLogger.WithField(key, value)
}

// WithFields adds multiple fields to the log entry
func WithFields(fields Fields) *apexlog.Entry {
	return DefaultLogger.WithFields(fields)
}
