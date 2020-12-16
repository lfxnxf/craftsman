package inits

import (
	"errors"
	"github.com/lfxnxf/craftsman/log"
	"github.com/lfxnxf/craftsman/opensearch"
)

type OpenSearchClients struct {
	logger  log.Logger
	clients map[string]*opensearch.Client
}

func (o *OpenSearchClients) GetClient(indexName string) (*opensearch.Client, error) {
	if o != nil {
		if client, ok := o.clients[indexName]; ok {
			return client, nil
		}
	}

	return nil, errors.New("index[" + indexName + "] no define")
}

func NewOpenSearchClients(logger log.Logger, tracerClients *TraceClients, openSearchConfigs []opensearch.OpenSearchConfig) (*OpenSearchClients, error) {
	OpenSearchClients := &OpenSearchClients{
		logger:  logger,
		clients: make(map[string]*opensearch.Client, len(openSearchConfigs)),
	}

	tracer, err := tracerClients.GetTracer()
	if err != nil {
		tracer = nil
	}

	for _, c := range openSearchConfigs {
		client := opensearch.NewClient(c, tracer, logger)
		OpenSearchClients.clients[c.IndexName] = client
	}

	return OpenSearchClients, nil
}
