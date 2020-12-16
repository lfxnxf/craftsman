package inits

import (
	"github.com/lfxnxf/craftsman/db/sql"
	"github.com/lfxnxf/craftsman/log"
)

type SQLClients struct {
	logger    log.Logger
	SQLGroups map[string]*sql.Group
}

func NewSQLClients(log log.Logger, tracerClients *TraceClients, sqlConfig []sql.SQLGroupConfig) (*SQLClients, error) {
	sqlGroup := &SQLClients{
		logger:    log,
		SQLGroups: make(map[string]*sql.Group, len(sqlConfig)),
	}

	tracer, err := tracerClients.GetTracer()
	if err != nil {
		tracer = nil
	}

	for _, c := range sqlConfig {
		g, err := sql.NewGroup(c, tracer)
		if err != nil {
			log.Error("new sql group", "config", c, "err", err)
			return sqlGroup, err
		}
		sqlGroup.SQLGroups[c.Name] = g
	}

	return sqlGroup, nil
}
