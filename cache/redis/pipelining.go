package redis

import (
	"context"
	"errors"
	"github.com/SkyAPM/go2sky/reporter/grpc/common"
	"github.com/tiantianjianbao/craftsman/tracing/sky"
	"sync"

	"github.com/garyburd/redigo/redis"
)

// Pipelining 提供了一些流水线的一些方法, 由NewPipelining函数创建
type Pipelining struct {
	conn    redis.Conn
	mu      sync.Mutex
	isClose bool
	ctx     context.Context
}

// NewPipelining函数创建一个Pipelining， 参数ctx用于trace系统
func (r *Redis) NewPipelining(ctx context.Context) (*Pipelining, error) {
	p := &Pipelining{}
	client := r.pool.Get()
	err := client.Err()
	if err != nil {
		return nil, err
	}
	p.conn = client
	p.mu = sync.Mutex{}
	if r.tracer != nil {
		span, err := r.tracer.CreateExitSpan(ctx, "NewPipelining", "NewPipelining", func(header string) error {
			return nil
		})
		if err == nil {
			span.SetSpanLayer(common.SpanLayer_Cache)
			r.ctx = sky.NewContext(ctx, span)
		}
	}

	return p, nil
}

func (p *Pipelining) Send(cmd string, args ...interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.isClose {
		return errors.New("Pipelining closed")
	}
	return p.conn.Send(cmd, args...)
}

func (p *Pipelining) Flush() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.isClose {
		return errors.New("Pipelining closed")
	}
	return p.conn.Flush()
}

func (p *Pipelining) Receive() (reply interface{}, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.isClose {
		return nil, errors.New("Pipelining closed")
	}
	return p.conn.Receive()
}

func (p *Pipelining) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.isClose = true
	span := sky.SpanFromContext(p.ctx)
	if span != nil {
		span.End()
	}

	return p.conn.Close()
}
