package zap

import (
	"bytes"
	"context"
	"io"

	"github.com/lfxnxf/craftsman/log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger...
type Logger struct {
	next log.Logger

	*zap.SugaredLogger

	path    string
	dir     string
	rolling RollingFormat

	rollingFiles []io.Writer

	loglevel zap.AtomicLevel
	prefix   string

	encoderCfg zapcore.EncoderConfig
	callSkip   int
}

var defaultEncoderConfig = zapcore.EncoderConfig{
	CallerKey:      "caller",
	StacktraceKey:  "stack",
	LineEnding:     zapcore.DefaultLineEnding,
	TimeKey:        "time",
	MessageKey:     "msg",
	LevelKey:       "level",
	NameKey:        "logger",
	EncodeCaller:   zapcore.ShortCallerEncoder,
	EncodeLevel:    zapcore.CapitalColorLevelEncoder,
	EncodeTime:     MilliSecondTimeEncoder,
	EncodeDuration: zapcore.StringDurationEncoder,
	EncodeName:     zapcore.FullNameEncoder,
}

//var logs = map[string]*Logger{}
var (
	LOG_ROTATE_HOUR  = "hour"
	LOG_ROTATE_DAY   = "day"
	LOG_ROTATE_MONTH = "month"

	LOG_LEVEL_DEBUG = "debug"
	LOG_LEVEL_INFO  = "info"
)

func NewJsonLogger(path, rotate, level string) (log.Logger, error) {
	cfg := defaultEncoderConfig
	cfg.EncodeLevel = zapcore.CapitalLevelEncoder

	lvl := zap.NewAtomicLevelAt(zap.DebugLevel)
	switch level {
	case LOG_LEVEL_DEBUG:
		lvl = zap.NewAtomicLevelAt(zap.DebugLevel)
	case LOG_LEVEL_INFO:
		lvl = zap.NewAtomicLevelAt(zap.InfoLevel)
	default:
		lvl = zap.NewAtomicLevelAt(zap.DebugLevel)
	}

	var rolling RollingFormat
	switch rotate {
	case LOG_ROTATE_HOUR:
		rolling = HourlyRolling
	case LOG_ROTATE_DAY:
		rolling = DailyRolling
	case LOG_ROTATE_MONTH:
		rolling = MinutelyRolling
	default:
		rolling = DailyRolling
	}

	rollFile, err := NewRollingFile(path, rolling)
	if err != nil {
		return nil, err
	}
	return &Logger{
		SugaredLogger: zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(cfg), rollFile, lvl)).WithOptions(zap.AddCaller(), zap.AddCallerSkip(1)).Sugar(),
		path:          path,
		dir:           "",
		rolling:       rolling,
		rollingFiles:  []io.Writer{rollFile},
		loglevel:      lvl,
		prefix:        "",
		encoderCfg:    cfg,
	}, nil
}

func (l Logger) Log(kvs ...interface{}) error {
	l.Infow("", kvs...)
	return nil
}

func (l Logger) DebugT(ctx context.Context, m string, kvs ...interface{}) {
	l.With(traceIDKey, traceID(ctx)).Debugw(m, kvs...)
}

func (l Logger) InfoT(ctx context.Context, m string, kvs ...interface{}) {
	l.With(traceIDKey, traceID(ctx)).Infow(m, kvs...)
}

func (l Logger) WarnT(ctx context.Context, m string, kvs ...interface{}) {
	l.With(traceIDKey, traceID(ctx)).Warnw(m, kvs...)
}

func (l Logger) ErrorT(ctx context.Context, m string, kvs ...interface{}) {
	l.With(traceIDKey, traceID(ctx)).Errorw(m, kvs...)
}

func (l Logger) PanicT(ctx context.Context, t string, kvs ...interface{}) {
	l.With(traceIDKey, traceID(ctx)).Panicw(t, kvs...)
}

func (l Logger) Debug(m string, args ...interface{}) {
	l.Debugw(m, args...)
}

func (l Logger) Info(m string, args ...interface{}) {
	l.Infow(m, args...)
}

func (l Logger) Warn(m string, args ...interface{}) {
	l.Warnw(m, args...)
}

func (l Logger) Error(m string, args ...interface{}) {
	l.Errorw(m, args...)
}

func (l Logger) Panic(m string, args ...interface{}) {
	l.Panicw(m, args...)
}

// Sync all log data
//func Sync() {
//	for _, l := range logs {
//		_ = l.Sync()
//	}
//	if _jsonDataLogger != nil {
//		_ = _jsonDataLogger.Sync()
//	}
//}

type logWriter struct {
	logFunc func() func(msg string, fileds ...interface{})
}

func (l logWriter) Write(p []byte) (int, error) {
	p = bytes.TrimSpace(p)
	if l.logFunc != nil {
		l.logFunc()(string(p))
	}
	return len(p), nil
}
