package processor

import (
	svcparameters "capyfile/capysvc/parameters"
	"capyfile/parameters"
	clientv3 "go.etcd.io/etcd/client/v3"
	"net/http"
)

type Context interface {
	ContextName() string
	ParameterLoaderProvider() (parameters.ParameterLoaderProvider, error)
}

type LocalContext struct {
}

func (processorContext *LocalContext) ContextName() string {
	return "local"
}

func (processorContext *LocalContext) ParameterLoaderProvider() (parameters.ParameterLoaderProvider, error) {
	return &svcparameters.GenericParameterLoaderProvider{}, nil
}

type UserSpaceContext struct {
}

func (processorContext *UserSpaceContext) ContextName() string {
	return "user_space"
}

func (processorContext *UserSpaceContext) ParameterLoaderProvider() (parameters.ParameterLoaderProvider, error) {
	return &svcparameters.GenericParameterLoaderProvider{}, nil
}

type HttpContext struct {
	Req *http.Request
}

func (processorContext *HttpContext) ContextName() string {
	return "http"
}

func (processorContext *HttpContext) ParameterLoaderProvider() (parameters.ParameterLoaderProvider, error) {
	return &svcparameters.GenericParameterLoaderProvider{Req: processorContext.Req}, nil
}

type DistKeyValueContext struct {
	EtcdClient *clientv3.Client
}

func (processorContext *DistKeyValueContext) ContextName() string {
	return "dist_kv"
}

func (processorContext *DistKeyValueContext) ParameterLoaderProvider() (parameters.ParameterLoaderProvider, error) {
	return &svcparameters.GenericParameterLoaderProvider{EtcdClient: processorContext.EtcdClient}, nil
}

type ChainContext struct {
	Chain []Context
}

func (processorContext *ChainContext) ContextName() string {
	return "chain"
}

func (processorContext *ChainContext) ParameterLoaderProvider() (parameters.ParameterLoaderProvider, error) {
	parameterLoaderProvider := &svcparameters.GenericParameterLoaderProvider{}

	for _, ctx := range processorContext.Chain {
		switch ctx.(type) {
		case *HttpContext:
			parameterLoaderProvider.Req = ctx.(*HttpContext).Req
			break
		case *DistKeyValueContext:
			parameterLoaderProvider.EtcdClient = ctx.(*DistKeyValueContext).EtcdClient
			break
		}
	}

	return parameterLoaderProvider, nil
}
