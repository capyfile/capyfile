package service

import (
	"capyfile/capysvc/processor"
	clientv3 "go.etcd.io/etcd/client/v3"
	"net/http"
)

// Context The context accessible to the service.
type Context interface {
	ContextName() string
	// ProcessorContext Builds processors context from the service context.
	ProcessorContext() processor.Context
}

type ServerContext struct {
	Req        *http.Request
	EtcdClient *clientv3.Client
}

func (serviceContext *ServerContext) ContextName() string {
	return "server"
}

func (serviceContext *ServerContext) ProcessorContext() processor.Context {
	ctx := &processor.ChainContext{
		Chain: []processor.Context{
			&processor.LocalContext{},
			&processor.UserSpaceContext{},
			&processor.HttpContext{Req: serviceContext.Req},
		},
	}
	if serviceContext.EtcdClient != nil {
		ctx.Chain = append(ctx.Chain, &processor.DistKeyValueContext{
			EtcdClient: serviceContext.EtcdClient,
		})
	}

	return ctx
}

type CliContext struct {
}

func (processorContext *CliContext) ContextName() string {
	return "cli"
}

func (processorContext *CliContext) ProcessorContext() processor.Context {
	return &processor.ChainContext{
		Chain: []processor.Context{
			&processor.LocalContext{},
			&processor.UserSpaceContext{},
		},
	}
}
