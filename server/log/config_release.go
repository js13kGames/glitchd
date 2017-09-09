// +build release

package log

import "go.uber.org/zap"

var Config = zap.Config{
	Level:         zap.NewAtomicLevel(),
	Development:   false,
	DisableCaller: true,
	Sampling: &zap.SamplingConfig{
		Initial:    100,
		Thereafter: 100,
	},
	Encoding:         "json",
	EncoderConfig:    zap.NewProductionEncoderConfig(),
	OutputPaths:      []string{"stderr"},
	ErrorOutputPaths: []string{"stderr"},
}
