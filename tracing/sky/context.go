package sky

import (
	"context"
	"github.com/SkyAPM/go2sky"
)

var defaultNoopSpan = &go2sky.NoopSpan{}

// SpanFromContext retrieves a go2sky Span from Go's context propagation
// mechanism if found. If not found, returns nil.
func SpanFromContext(ctx context.Context) go2sky.Span {
	if ctx != nil {
		if s, ok := ctx.Value(spanKey).(go2sky.Span); ok {
			return s
		}
	}

	return nil
}

// SpanOrNoopFromContext retrieves a Zipkin Span from Go's context propagation
// mechanism if found. If not found, returns a noopSpan.
// This function typically is used for modules that want to provide existing
// go2sky spans with additional data, but can't guarantee that spans are
// properly propagated. It is preferred to use SpanFromContext() and test for
// Nil instead of using this function.
func SpanOrNoopFromContext(ctx context.Context) go2sky.Span {
	if s, ok := ctx.Value(spanKey).(go2sky.Span); ok {
		return s
	}
	return defaultNoopSpan
}

// NewContext stores a go2sky Span into Go's context propagation mechanism.
func NewContext(ctx context.Context, s go2sky.Span) context.Context {
	return context.WithValue(ctx, spanKey, s)
}

type ctxKey struct{}

var spanKey = ctxKey{}
