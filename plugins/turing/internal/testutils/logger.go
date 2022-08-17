package testutils

import (
	"bytes"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/caraml-dev/turing/engines/experiment/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// MemorySink implements zap.Sink by writing all messages to a buffer.
type MemorySink struct {
	*bytes.Buffer
}

// Close is a nop method to satisfy the zap.Sink interface
func (s *MemorySink) Close() error { return nil }

// Sync is a nop method to satisfy the zap.Sink interface
func (s *MemorySink) Sync() error { return nil }

// NewLoggerWithMemorySink creates a new logger with a memory sink, to which all
// output is redirected. Calling sink.String() / sink.Bytes() ... will give access to
// its contents.
func NewLoggerWithMemorySink() (log.Logger, *MemorySink, error) {
	sinkName := fmt.Sprintf("memory-%s", strconv.FormatInt(time.Now().UnixNano(), 10))
	memorySink := &MemorySink{new(bytes.Buffer)}

	_ = zap.RegisterSink(sinkName, func(*url.URL) (zap.Sink, error) {
		return memorySink, nil
	})

	// Using the default prod config with Debug level, set the memory sink as the output
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	cfg.OutputPaths = []string{fmt.Sprintf("%s://", sinkName)}

	logger, err := cfg.Build()
	if err != nil {
		return nil, nil, err
	}
	return &zapLogger{SugaredLogger: logger.Sugar(), cfg: &cfg}, memorySink, nil
}

type zapLogger struct {
	*zap.SugaredLogger
	cfg *zap.Config
}

func (l *zapLogger) With(args ...interface{}) log.Logger {
	return &zapLogger{l.SugaredLogger.With(args...), l.cfg}
}

func (l *zapLogger) SetLevel(lvl string) {
	var zapLvl zapcore.Level
	if err := zapLvl.UnmarshalText([]byte(lvl)); err != nil {
		l.Warnf("failed to set %s log level: %v", lvl, err)
	} else {
		l.cfg.Level = zap.NewAtomicLevelAt(zapLvl)
	}
}
