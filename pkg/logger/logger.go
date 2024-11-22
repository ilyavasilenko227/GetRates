package logger

import (
	"os"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	LevelDebug = "DEBUG"
	LevelInfo  = "INFO"
)

var (
	logger      *zap.Logger
	once        sync.Once
	atomicLevel zap.AtomicLevel
)

func BuildLogger(logLevel string) {
	once.Do(func() {
		atomicLevel = zap.NewAtomicLevel()
		SetLevel(logLevel)
		encoderCfg := zap.NewProductionEncoderConfig()
		logger = zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg), os.Stdout, atomicLevel), zap.AddCaller())
	})
}

func SetLevel(logLevel string) {
	switch strings.ToUpper(logLevel) {
	case LevelDebug:
		atomicLevel.SetLevel(zapcore.DebugLevel)
	case LevelInfo:
		atomicLevel.SetLevel(zapcore.InfoLevel)
	default:
		panic("invalid log level specified for logger")
	}
}

func CurrentLevel() string {
	return atomicLevel.String()
}

func Logger() *zap.Logger {
	if logger == nil {
		BuildLogger(LevelDebug)
	}
	return logger
}
