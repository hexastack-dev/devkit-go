package zaplog

import (
	"io"

	"github.com/hexastack-dev/devkit-go/log"
)

type Encoder uint8

const (
	JSONEncoder Encoder = iota
	ConsoleEncoder
)

// FileLogConfig define configurations for rolling file log.
type FileLogConfig struct {
	// Enabled wether to log to a file or not.
	Enabled bool
	// LogLevel to use when logging to a file, any log lower than this will not be logged.
	// Default to use InfoLogLevel
	LogLevel log.LogLevel
	// Encoder to use when loging to file. Default to JSONEncoder.
	Encoder Encoder
	// Filename is the file to write logs to.  Backup log files will be retained in the same directory.
	Filename string
	// MaxSize is the maximum size in megabytes of the log file before it gets rotated. Default to 100 megabytes.
	MaxSize int
	// MaxBackups is the maximum number of old log files to retain. The default is to retain all old log files.
	MaxBackups int
}

// Config is configuration for log. Any underlying log framework should comply with this spec.
// By defaults this configuration specifies console logger. See FileLogConfig for file based log configurations.
// Console log cannot be disabled.
type Config struct {
	// RootLogLevel define root log level to use, any log lower than this will not be logged.
	// Default to InfoLogLevel.
	RootLogLevel log.LogLevel
	// Encoder to use to log default to ConsoleEncoder.
	Encoder Encoder
	// FileLogConfig configuration for rolling file log.
	FileLogConfig FileLogConfig
	// Output set log output for console. Defaults to os.Stdout.
	Output io.Writer
}
