// +build !release

package log

import "go.uber.org/zap"

var Config = zap.Config{
	Level:            zap.NewAtomicLevel(),
	Development:      true,
	DisableCaller:    false,
	Encoding:         "console",
	EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
	OutputPaths:      []string{"stderr"},
	ErrorOutputPaths: []string{"stderr"},
}
