package log

import "context"

type Logger interface {
	Log(kvs ...interface{}) error

	Debug(m string, kvs ...interface{})
	Info(m string, kvs ...interface{})
	Warn(m string, kvs ...interface{})
	Error(m string, kvs ...interface{})
	Panic(m string, kvs ...interface{})

	DebugT(ctx context.Context, m string, kvs ...interface{})
	InfoT(ctx context.Context, m string, kvs ...interface{})
	WarnT(ctx context.Context, m string, kvs ...interface{})
	ErrorT(ctx context.Context, m string, kvs ...interface{})
	PanicT(ctx context.Context, m string, kvs ...interface{})
}
