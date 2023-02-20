package log_test

import (
	"errors"
	"testing"

	"github.com/hexastack-dev/devkit-go/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setGlobalLogger(observer *logObserver) {
	logger := log.NewSimpleLogger(observer, log.DebugLogLevel)
	log.SetLogger(logger)
}

func TestDebug(t *testing.T) {
	observer := &logObserver{}
	setGlobalLogger(observer)

	writeGlobalLog(log.DebugLogLevel)
	assert.Equal(t, 1, len(observer.entries))
	assert.Greater(t, len(observer.entries[0]), 40)
	suf := observer.entries[0][39:]
	assert.Equal(t, "level:debug\tmessage:Hello\n", suf)
}

func TestInfo(t *testing.T) {
	observer := &logObserver{}
	setGlobalLogger(observer)

	writeGlobalLog(log.InfoLogLevel)
	require.Equal(t, 1, len(observer.entries))
	require.Greater(t, len(observer.entries[0]), 40)
	suf := observer.entries[0][39:]
	assert.Equal(t, "level:info\tmessage:Hello\n", suf)
}

func TestWarn(t *testing.T) {
	observer := &logObserver{}
	setGlobalLogger(observer)

	writeGlobalLog(log.WarnLogLevel)
	assert.Equal(t, 1, len(observer.entries))
	assert.Greater(t, len(observer.entries[0]), 40)
	suf := observer.entries[0][39:]
	assert.Equal(t, "level:warn\tmessage:Hello\n", suf)
}

func TestError(t *testing.T) {
	observer := &logObserver{}
	setGlobalLogger(observer)

	writeGlobalErrorLog(errors.New("oopsie"))
	assert.Equal(t, 1, len(observer.entries))
	assert.Greater(t, len(observer.entries[0]), 40)
	suf := observer.entries[0][39:]
	assert.Equal(t, "level:error\terror:oopsie\tmessage:Something went wrong\n", suf)
}

func writeGlobalLog(lv log.LogLevel, fields ...log.LogField) {
	switch lv {
	case log.DebugLogLevel:
		log.Debug("Hello", fields...)
	case log.InfoLogLevel:
		log.Info("Hello", fields...)
	case log.WarnLogLevel:
		log.Warn("Hello", fields...)
	}
}

func writeGlobalErrorLog(err error, fields ...log.LogField) {
	log.Error("Something went wrong", err, fields...)
}
