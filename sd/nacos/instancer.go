package nacos

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/util/conn"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/nacos-group/nacos-sdk-go/model"
)

const defaultIndex = 0

// errStopped notifies the loop to quit. aka stopped via quitc
var errStopped = errors.New("quit and closed consul instancer")

// Instancer yields instances for a service in Consul.
type Instancer struct {
	cache       *Cache
	client      Client
	logger      log.Logger
	service     string
	tags        []string
	passingOnly bool
	quitc       chan struct{}
}



// NewInstancer returns a Consul instancer that publishes instances for the
// requested service. It only returns instances for which all of the passed tags
// are present.
func NewInstancer(client Client, logger log.Logger, service string, tags []string, passingOnly bool) *Instancer {
	s := &Instancer{
		cache:       NewCache(),
		client:      client,
		logger:      log.With(logger, "service", service, "tags", fmt.Sprint(tags)),
		service:     service,
		tags:        tags,
		passingOnly: passingOnly,
		quitc:       make(chan struct{}),
	}

	instances, index, err := s.getInstances(defaultIndex, nil)
	if err == nil {
		s.logger.Log("instances", len(instances))
	} else {
		s.logger.Log("err", err)
	}

	s.cache.Update(sd.Event{Instances: instances, Err: err})
	go s.loop(index)
	return s
}

func NewInstancer2(client Client, logger log.Logger, service string, tags []string, passingOnly bool) *Instancer {
	s := &Instancer{
		cache:       NewCache(),
		client:      client,
		logger:      log.With(logger, "service", service, "tags", fmt.Sprint(tags)),
		service:     service,
		tags:        tags,
		passingOnly: passingOnly,
		quitc:       make(chan struct{}),
	}

	instances, _, err := s.getInstances(defaultIndex, nil)
	if err == nil {
		s.logger.Log("instances", len(instances))
	} else {
		s.logger.Log("err", err)
	}
	return s
}

// Stop terminates the instancer.
func (s *Instancer) Stop() {
	close(s.quitc)
}

func (s *Instancer) loop(lastIndex uint64) {
	var (
		instances []string
		err       error
		//d         time.Duration = 10 * time.Millisecond
		d         time.Duration = 10 * time.Second
	)
	for {
		instances, lastIndex, err = s.getInstances(lastIndex, s.quitc)
		switch {
		case err == errStopped:
			return // stopped via quitc
		case err != nil:
			s.logger.Log("err", err)
			time.Sleep(d)
			d = conn.Exponential(d)
			s.cache.Update(sd.Event{Err: err})
		default:
			fmt.Println(instances , "xxx")
			s.cache.Update(sd.Event{Instances: instances})
			//d = 10 * time.Millisecond
			d = 5 * time.Second

			// TODO ----------
			time.Sleep(d)

		}
	}
}

func (s *Instancer) getInstances(lastIndex uint64, interruptc chan struct{}) ([]string, uint64, error) {
	type response struct {
		instances []string
		index     uint64
	}

	var (
		errc = make(chan error, 1)
		resc = make(chan response, 1)
	)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				_ = s.logger.Log("err", err)
			}
		}()
		entries, err := s.client.SelectInstances(vo.SelectInstancesParam{
			ServiceName: s.service,
			Clusters:    []string{"a"},
			GroupName:   "",
			HealthyOnly: true,
		})
		if err != nil {
			errc <- err
			return
		}
		resc <- response{
			instances: makeInstances(entries),
		}
	}()

	select {
	case err := <-errc:
		return nil, 0, err
	case res := <-resc:
		return res.instances, res.index, nil
	case <-interruptc:
		return nil, 0, errStopped
	}
}

// Register implements Instancer.
func (s *Instancer) Register(ch chan<- sd.Event) {
	s.cache.Register(ch)
}

// Deregister implements Instancer.
func (s *Instancer) Deregister(ch chan<- sd.Event) {
	s.cache.Deregister(ch)
}

func makeInstances(entries []model.Instance) []string {
	instances := make([]string, len(entries))
	for i, entry := range entries {
		instances[i] = fmt.Sprintf("%s:%d", entry.Ip, entry.Port)
	}
	return instances
}
