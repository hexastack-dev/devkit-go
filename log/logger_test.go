package log_test

import (
	"errors"
	"testing"

	"github.com/hexastack-dev/devkit-go/log"
)

func BenchmarkLogger(b *testing.B) {
	logger := &log.NoOpLogger{}

	b.ResetTimer()
	b.Run("log", func(b *testing.B) {
		testLog(b, logger)
	})

	b.ResetTimer()
	b.Run("log with 10 fields", func(b *testing.B) {
		testLogWithArguments(b, logger)
	})

	b.ResetTimer()
	b.Run("log error with 10 fields", func(b *testing.B) {
		testError(b, logger)
	})
}

func testLog(b *testing.B, logger log.Logger) {
	for i := 0; i < b.N; i++ {
		logger.Info("info message")
	}
}

func testLogWithArguments(b *testing.B, logger log.Logger) {
	for i := 0; i < b.N; i++ {
		logger.Info("info message", generateField(i)...)
	}
}

func testError(b *testing.B, logger log.Logger) {
	for i := 0; i < b.N; i++ {
		err := errors.New("oops")
		logger.Error("something went wrong", err, generateField(i)...)
	}
}

func generateField(i int) []log.LogField {
	return []log.LogField{
		log.Field("a", "1"),
		log.Field("b", "2"),
		log.Field("c", "3"),
		log.Field("d", "4"),
		log.Field("e", "5"),
		log.Field("f", "s1"),
		log.Field("g", "s2"),
		log.Field("h", "s3"),
		log.Field("i", "s4"),
		log.Field("j", "s5"),
	}
}
