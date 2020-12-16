package tracing

import (
	"errors"
	"math/rand"

	"github.com/SkyAPM/go2sky"
)

// Balancer yields endpoints according to some heuristic.
type Balancer interface {
	Endpoint() (*go2sky.Tracer, error)
}

// ErrNoEndpoints is returned when no qualifying endpoints are available.
var ErrNoEndpoints = errors.New("no endpoints available")

type Endpointer interface {
	Endpoints() ([]*go2sky.Tracer, error)
}

type FixedEndpointer []*go2sky.Tracer

// Endpoints implements Endpointer.
func (s FixedEndpointer) Endpoints() ([]*go2sky.Tracer, error) { return s, nil }

// NewRandom returns a load balancer that selects services randomly.
func NewRandom(s Endpointer, seed int64) Balancer {
	return &random{
		s: s,
		r: *rand.New(rand.NewSource(seed)),
	}
}

type random struct {
	s Endpointer
	r rand.Rand
}

func (r *random) Endpoint() (*go2sky.Tracer, error) {
	endpoints, err := r.s.Endpoints()
	if err != nil {
		return &go2sky.Tracer{}, err
	}
	if len(endpoints) <= 0 {
		return &go2sky.Tracer{}, ErrNoEndpoints
	}
	return endpoints[r.r.Intn(len(endpoints))], nil
}
