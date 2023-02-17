package zaplog

import (
	"context"
	// "github.com/uptrace/opentelemetry-go-extra/otelzap"
	"os"
	"path/filepath"

	"go.opentelemetry.io/otel/trace"

	"github.com/hexastack-dev/devkit-go/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// New create new instance of Logger.
func New(zapLogger *zap.Logger) *Logger {
	return &Logger{
		zlog: zapLogger,
	}
}

// NewDefaultLogger create new instance of Zap Logger using
// default configuration.
func NewDefaultLogger(config Config) *Logger {
	if config.Output == nil {
		config.Output = os.Stdout
	}
	return &Logger{
		zlog: configureZap(config),
	}
}

func configureZap(config Config) *zap.Logger {
	conf := zap.NewProductionEncoderConfig()
	conf.TimeKey = "timestamp"
	conf.EncodeTime = zapcore.ISO8601TimeEncoder
	conf.MessageKey = "message"

	outputs := configureOutputs(config, conf)
	core := zapcore.NewTee(outputs...)
	return zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
}

func configureOutputs(config Config, enconfig zapcore.EncoderConfig) []zapcore.Core {
	cores := []zapcore.Core{
		zapcore.NewCore(
			buildZapEncoder(config.Encoder, enconfig),
			zapcore.AddSync(config.Output),
			mapLogLevel(config.RootLogLevel)),
	}
	if config.FileLogConfig.Enabled {
		cores = append(cores, zapcore.NewCore(
			buildZapEncoder(config.FileLogConfig.Encoder, enconfig),
			zapcore.AddSync(rollingFile(config)),
			mapLogLevel(config.FileLogConfig.LogLevel)),
		)
	}
	return cores
}

func buildZapEncoder(encoder Encoder, enconfig zapcore.EncoderConfig) zapcore.Encoder {
	switch encoder {
	case ConsoleEncoder:
		return zapcore.NewConsoleEncoder(enconfig)
	default:
		return zapcore.NewJSONEncoder(enconfig)
	}
}

func mapLogLevel(level log.LogLevel) zap.AtomicLevel {
	switch level {
	case log.FatalLogLevel:
		return zap.NewAtomicLevelAt(zap.FatalLevel)
	case log.ErrorLogLevel:
		return zap.NewAtomicLevelAt(zap.ErrorLevel)
	case log.WarnLogLevel:
		return zap.NewAtomicLevelAt(zap.WarnLevel)
	case log.DebugLogLevel:
		return zap.NewAtomicLevelAt(zap.DebugLevel)
	default:
		return zap.NewAtomicLevelAt(zap.InfoLevel)
	}
}

// rollingFile create lumberjack.Logger instance, will return nil if
// config.FilelogEnabled is false, and will panic if the log path
// cannot be resolved.
func rollingFile(config Config) *lumberjack.Logger {
	path := config.FileLogConfig.Filename
	if path == "" {
		path = filepath.Base(os.Args[0]) + ".log"
	}
	path, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	return &lumberjack.Logger{
		Filename:   path,
		MaxBackups: config.FileLogConfig.MaxBackups,
		MaxSize:    config.FileLogConfig.MaxSize,
	}
}

var _ log.Logger = &Logger{}

type Logger struct {
	zlog *zap.Logger
	ctx  context.Context
	// otelog *otelzap.Logger
}

// Fatal logs a message at FatalLevel, then calls os.Exit(1). This should only be use with extra care, ideally Fatal
// should only be used in main where the apps encountered an error and have noway to continue.
// optfields is optional, when supplied it will be added as new field using
// Key as field name, and Value as it's value.
func (l *Logger) Fatal(msg string, err error, optfields ...log.LogField) {
	zfields := make([]zap.Field, 0, len(optfields))
	zfields = append(zfields, zap.Error(err))
	zfields = append(zfields, convertFields(optfields)...)
	if l.ctx != nil {
		zfields = append(zfields, fromContext(l.ctx)...)
		// l.otelog.Ctx(l.ctx).Fatal(msg, zfields...)
		// return
	}

	l.zlog.Fatal(msg, zfields...)
}

// Error logs a message at ErrorLevel, put the passed error in "error" field.
// optfields is optional, when supplied it will be added as new field using
// Key as field name, and Value as it's value.
func (l *Logger) Error(msg string, err error, optfields ...log.LogField) {
	zfields := make([]zap.Field, 0, len(optfields))
	zfields = append(zfields, zap.Error(err))
	zfields = append(zfields, convertFields(optfields)...)
	if l.ctx != nil {
		zfields = append(zfields, fromContext(l.ctx)...)
		// l.otelog.Ctx(l.ctx).Error(msg, zfields...)
		// return
	}

	l.zlog.Error(msg, zfields...)
}

// Warn logs a message at WarnLevel.
// optfields is optional, when supplied it will be added as new field using
// Key as field name, and Value as it's value.
func (l *Logger) Warn(msg string, optfields ...log.LogField) {
	zfields := make([]zap.Field, 0, len(optfields))
	zfields = append(zfields, convertFields(optfields)...)
	if l.ctx != nil {
		zfields = append(zfields, fromContext(l.ctx)...)
		// l.otelog.Ctx(l.ctx).Warn(msg, zfields...)
		// return
	}

	l.zlog.Warn(msg, zfields...)
}

// Info logs a message at InfoLevel.
// optfields is optional, when supplied it will be added as new field using
// Key as field name, and Value as it's value.
func (l *Logger) Info(msg string, optfields ...log.LogField) {
	zfields := make([]zap.Field, 0, len(optfields))
	zfields = append(zfields, convertFields(optfields)...)
	if l.ctx != nil {
		zfields = append(zfields, fromContext(l.ctx)...)
		// l.otelog.Ctx(l.ctx).Info(msg, zfields...)
		// return
	}

	l.zlog.Info(msg, zfields...)
}

// Debug logs a message at DebugLevel.
// optfields is optional, when supplied it will be added as new field using
// Key as field name, and Value as it's value.
func (l *Logger) Debug(msg string, optfields ...log.LogField) {
	zfields := make([]zap.Field, 0, len(optfields))
	zfields = append(zfields, convertFields(optfields)...)
	if l.ctx != nil {
		zfields = append(zfields, fromContext(l.ctx)...)
		// l.otelog.Ctx(l.ctx).Debug(msg, zfields...)
		// return
	}

	l.zlog.Debug(msg, zfields...)
}

// WithContext return Logger instance that will use passed context to log additional info,
// such as opentelemetry's SpanID and TraceID if applicable.
func (l *Logger) WithContext(ctx context.Context) log.Logger {
	return &Logger{
		zlog: l.zlog,
		ctx:  ctx,
		// otelog: otelzap.New(l.zlog, otelzap.WithMinLevel(zapcore.InfoLevel)),
	}
}

// Sync will calls zap logger Sync(), this method should be called
// before the program exit.
//
//	ie. func main() {
//		...
//		zlog := New(conf)
//		defer zlog.Sync()
//		...
//	}
func (l *Logger) Sync() error {
	return l.zlog.Sync()
}

func (l *Logger) WithOptions(opts ...zap.Option) *Logger {
	zlog := l.zlog.WithOptions(opts...)
	return &Logger{
		zlog: zlog,
		ctx:  l.ctx,
	}
}

func fromContext(ctx context.Context) []zap.Field {
	otelFields := make([]zap.Field, 0)
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return otelFields
	}
	sc := span.SpanContext()
	if sc.HasSpanID() {
		otelFields = append(otelFields, zap.String("spanId", sc.SpanID().String()))
	}
	if sc.HasTraceID() {
		otelFields = append(otelFields, zap.String("traceId", sc.TraceID().String()))
	}
	otelFields = append(otelFields, zap.Int("traceFlags", int(sc.TraceFlags())))
	return otelFields
}

func convertFields(fields []log.LogField) []zapcore.Field {
	var zfields []zapcore.Field

	for _, field := range fields {
		zfields = append(zfields, zap.Any(field.Key, field.Value))
	}

	return zfields
}
