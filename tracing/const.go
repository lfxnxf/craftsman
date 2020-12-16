package tracing

const (
	TagURL             = "url"
	TagStatusCode      = "status_code"
	TagHTTPMethod      = "http.method"
	TagHTTPClient      = "http.client"
	TagHTTPServer      = "http.server"
	TagGRPCClientStart = "grpc.client.start"
	TagGRPCClientEnd   = "grpc.client.end"
	TagGRPCClient      = "grpc.client.end"
	TagGRPCServer      = "grpc.server"
	TagGRPCServerStart = "grpc.server.start"
	TagGRPCServerEnd   = "grpc.server.end"
	TagGRPCMethod      = "grpc.method"
	TagDBType          = "db.type"
	TagOpenSearchType  = "opensearch.type"
	TagDBInstance      = "db.instance"
	TagDBStatement     = "db.statement"
	TagDBBindVariables = "db.bind_vars"
	TagMQQueue         = "mq.queue"
	TagMQBroker        = "mq.broker"
	TagMQTopic         = "mq.topic"
	TagMQID            = "mq.id"
)

const (
	PropagationHeader = "propagation_header"
	TraceId           = "trace_id"
)
