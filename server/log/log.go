package log

import "go.uber.org/zap"

func New() (*zap.Logger, error) {
	return Config.Build()
}

func SetLevel(level string) {
	switch level {
	case "panic":
		Config.Level.SetLevel(zap.PanicLevel)
	case "fatal":
		Config.Level.SetLevel(zap.FatalLevel)
	case "error":
		Config.Level.SetLevel(zap.ErrorLevel)
	case "warn", "warning":
		Config.Level.SetLevel(zap.WarnLevel)
	case "info":
		Config.Level.SetLevel(zap.InfoLevel)
	case "debug":
		Config.Level.SetLevel(zap.DebugLevel)
	}
}
