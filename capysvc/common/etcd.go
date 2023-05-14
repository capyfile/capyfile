package common

import (
	"encoding/json"
	clientv3 "go.etcd.io/etcd/client/v3"
	"os"
)

// EtcdClient Etcd is optional here, and it's used only to load parameters.
var EtcdClient *clientv3.Client

func InitEtcdClient() error {
	// env ETCD_ENDPOINTS=["etcd:2379", "etcd:22379"]
	etcdEndpointEnvVar, ok := os.LookupEnv("ETCD_ENDPOINTS")
	if ok {
		var endpoints []string
		endpointsErr := json.Unmarshal([]byte(etcdEndpointEnvVar), &endpoints)
		if endpointsErr != nil {
			return endpointsErr
		}

		username, _ := os.LookupEnv("ETCD_USERNAME")
		password, _ := os.LookupEnv("ETCD_PASSWORD")

		client, clientErr := clientv3.New(clientv3.Config{
			Endpoints: endpoints,
			Username:  username,
			Password:  password,
		})
		if clientErr != nil {
			return clientErr
		}

		EtcdClient = client
	}

	return nil
}
