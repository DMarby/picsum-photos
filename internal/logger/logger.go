package logger

import (
	stdlog "log"
	"os"
	"strings"

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
		return lvl >= loglevel && lvl >= zapcore.ErrorLevel
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
	log := zap.New(core, zap.AddCaller())

	// Redirect stdlib log package to zap
	_, _ = zap.RedirectStdLogAt(log, zapcore.ErrorLevel)

	return &Logger{
		log.Sugar(),
	}
}

type httpErrorLog struct {
	log *Logger
}

func (h *httpErrorLog) Write(p []byte) (int, error) {
	m := string(p)

	if strings.HasPrefix(m, "http: URL query contains semicolon") {
		h.log.Debug(m)
	} else {
		h.log.Error(m)
	}

	return len(p), nil
}

func NewHTTPErrorLog(logger *Logger) *stdlog.Logger {
	return stdlog.New(&httpErrorLog{logger}, "", 0)
}
