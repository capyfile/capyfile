package parameters

import (
	"capyfile/parameters"
	clientv3 "go.etcd.io/etcd/client/v3"
	"net/http"
)

type GenericParameterLoaderProvider struct {
	// Req The request instance is required to load parameters from "http_*" source types.
	Req *http.Request
	// EtcdClient The etcd client is required to load parameters from "etcd" source type.
	EtcdClient *clientv3.Client
}

func (provider *GenericParameterLoaderProvider) HasParameterLoader(sourceType string) bool {
	switch sourceType {
	case "value":
	case "env_var":
	case "file":
	case "secret":
		return true
	case "http_get":
	case "http_post":
	case "http_header":
		return provider.Req != nil
	case "etcd":
		return provider.EtcdClient != nil
	}

	return false
}

func (provider *GenericParameterLoaderProvider) ParameterLoader(
	sourceType string,
	source any,
) (parameters.ParameterLoader, error) {
	return &genericParameterLoader{
		sourceType: sourceType,
		source:     source,
		req:        provider.Req,
		etcdClient: provider.EtcdClient,
	}, nil
}

type genericParameterLoader struct {
	sourceType string
	source     any

	req        *http.Request
	etcdClient *clientv3.Client
}

func (parameterLoader *genericParameterLoader) LoadBoolValue() (bool, error) {
	return retrieveBoolParameterValue(
		parameterLoader.sourceType,
		parameterLoader.source,
		parameterLoader.req,
		parameterLoader.etcdClient,
	)
}

func (parameterLoader *genericParameterLoader) LoadIntValue() (int64, error) {
	return retrieveIntParameterValue(
		parameterLoader.sourceType,
		parameterLoader.source,
		parameterLoader.req,
		parameterLoader.etcdClient,
	)
}

func (parameterLoader *genericParameterLoader) LoadStringValue() (string, error) {
	return retrieveStringParameterValue(
		parameterLoader.sourceType,
		parameterLoader.source,
		parameterLoader.req,
		parameterLoader.etcdClient,
	)
}

func (parameterLoader *genericParameterLoader) LoadStringArrayValue() ([]string, error) {
	return retrieveStringArrayParameterValue(
		parameterLoader.sourceType,
		parameterLoader.source,
		parameterLoader.req,
		parameterLoader.etcdClient,
	)
}
