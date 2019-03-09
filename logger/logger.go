package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is a logger
type Logger struct {
	*zap.SugaredLogger
}

// New creates a new logger
func New(loglevel zapcore.Level) *Logger {
	// Configure console output.
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	consoleEncoder := zapcore.NewJSONEncoder(encoderConfig)

	// Log errors to stderr
	stderrLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})

	stdoutLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= loglevel && lvl < zapcore.ErrorLevel
	})
	stdout := zapcore.Lock(os.Stdout)
	stderr := zapcore.Lock(os.Stderr)

	// Merge the outputs, encoders, and level-handling functions
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, stderr, stderrLevel),
		zapcore.NewCore(consoleEncoder, stdout, stdoutLevel),
	)

	// Construct our logger
	log := zap.New(core)

	// Redirect stdlib log package to zap
	zap.RedirectStdLog(log)

	return &Logger{
		log.Sugar(),
	}
}
