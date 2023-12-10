package capysvc

import (
	svcparameters "capyfile/capysvc/parameters"
	"capyfile/parameters"
	clientv3 "go.etcd.io/etcd/client/v3"
	"net/http"
)

type Context interface {
	Request() *http.Request
	EtcdClient() *clientv3.Client
	ParameterLoaderProvider() (parameters.ParameterLoaderProvider, error)
}

type ServerContext struct {
	req        *http.Request
	etcdClient *clientv3.Client
}

func NewServerContext(req *http.Request, etcdClient *clientv3.Client) *ServerContext {
	return &ServerContext{
		req:        req,
		etcdClient: etcdClient,
	}
}

func (c *ServerContext) Request() *http.Request {
	return c.req
}

func (c *ServerContext) EtcdClient() *clientv3.Client {
	return c.etcdClient
}

func (c *ServerContext) ParameterLoaderProvider() (parameters.ParameterLoaderProvider, error) {
	return &svcparameters.GenericParameterLoaderProvider{
		Req:        c.req,
		EtcdClient: c.etcdClient,
	}, nil
}

type CliContext struct{}

func NewCliContext() *CliContext {
	return &CliContext{}
}

func (c *CliContext) Request() *http.Request {
	return nil
}

func (c *CliContext) EtcdClient() *clientv3.Client {
	return nil
}

func (c *CliContext) ParameterLoaderProvider() (parameters.ParameterLoaderProvider, error) {
	return &svcparameters.GenericParameterLoaderProvider{}, nil
}

type WorkerContext struct {
	etcdClient *clientv3.Client
}

func NewWorkerContext(etcdClient *clientv3.Client) *WorkerContext {
	return &WorkerContext{
		etcdClient: etcdClient,
	}
}

func (c *WorkerContext) Request() *http.Request {
	return nil
}

func (c *WorkerContext) EtcdClient() *clientv3.Client {
	return c.etcdClient
}

func (c *WorkerContext) ParameterLoaderProvider() (parameters.ParameterLoaderProvider, error) {
	return &svcparameters.GenericParameterLoaderProvider{
		EtcdClient: c.etcdClient,
	}, nil
}
