package zaplog_test

import (
	"github.com/hexastack-dev/devkit-go/log"
	"github.com/hexastack-dev/devkit-go/log/drivers/zaplog"
)

func ExampleNew() {
	config := zaplog.Config{
		RootLogLevel: log.DebugLogLevel,
		FileLogConfig: zaplog.FileLogConfig{
			Enabled:  true,
			Filename: "./log/app.log",
		},
	}
	zlog := zaplog.NewDefaultLogger(config)
	defer zlog.Sync()

	zlog.Info("Hello")

	v1 := log.Field("v1", "V1")
	zlog.Info("World", v1)
}
