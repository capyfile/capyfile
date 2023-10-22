package parameters

import (
	"capyfile/capysvc/common"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/exp/slog"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"time"
)

// retrieveStringArrayParameterValue Anything that is more complex than scalar value should be JSON, at least for now.
func retrieveStringArrayParameterValue(
	valueType string,
	value any,
	request *http.Request,
	etcdClient *clientv3.Client,
) ([]string, error) {
	if valueType == "value" {
		return anyToStringArrayValue(value), nil
	}

	var rawValue string
	switch valueType {
	case "http_get":
		rawValue = request.URL.Query().Get(value.(string))
		if len(rawValue) == 0 {
			return []string{}, errors.New(
				fmt.Sprintf(
					"request get parameter \"%s\" can not be found. Please make sure it is set",
					value.(string)))
		}
		break
	case "http_post":
		err := request.ParseForm()
		if err != nil {
			return []string{}, errors.New("request post parameters can not be parsed")
		}
		rawValue = request.Form.Get(value.(string))
		if len(rawValue) == 0 {
			return []string{}, errors.New(
				fmt.Sprintf(
					"request post parameter \"%s\" can not be found. Please make sure it is set",
					value.(string)))
		}
		break
	case "http_header":
		rawValue = request.Header.Get(value.(string))
		if len(rawValue) == 0 {
			return []string{}, errors.New(
				fmt.Sprintf(
					"request header parameter \"%s\" can not be found. Please make sure it is set",
					value.(string)))
		}
		break
	case "env_var":
		v, found := os.LookupEnv(value.(string))
		if !found {
			return []string{}, errors.New(
				fmt.Sprintf(
					"environment variable \"%s\" can not be found. Please make sure it is set",
					value.(string)))
		}
		rawValue = v
		break
	case "secret":
		b, err := os.ReadFile("/run/secrets/" + value.(string))
		if err != nil {
			return []string{}, errors.New(
				fmt.Sprintf(
					"secret \"%s\" can not be read. Please make sure it is mounted",
					value.(string)))
		}
		rawValue = string(b)
		break
	case "file":
		b, err := os.ReadFile(value.(string))
		if err != nil {
			return []string{}, errors.New(
				fmt.Sprintf("file \"%s\" can not be read. Please make sure it is redable", value.(string)))
		}
		rawValue = string(b)
		break
	case "etcd":
		kv, kvErr := etcdKeyValue(etcdClient, value.(string))
		if kvErr != nil {
			return []string{}, fmt.Errorf("etcd key \"%s\" can not be read", value.(string))
		}

		rawValue = kv.String()

		break
	}

	if len(rawValue) == 0 {
		return []string{}, errors.New("failed to retrieve string array parameter")
	}

	var parameterValue []string
	err := json.Unmarshal([]byte(rawValue), &parameterValue)

	return parameterValue, err
}

func retrieveStringParameterValue(
	valueType string,
	value any,
	request *http.Request,
	etcdClient *clientv3.Client,
) (string, error) {
	switch valueType {
	case "value":
		return anyToStringValue(value)
	case "http_get":
		v := request.URL.Query().Get(value.(string))
		if len(v) == 0 {
			return "", errors.New(
				fmt.Sprintf(
					"request get parameter \"%s\" can not be found. Please make sure it is set",
					value.(string)))
		}
		return v, nil
	case "http_post":
		err := request.ParseForm()
		if err != nil {
			return "", errors.New("request post parameters can not be parsed")
		}

		v := request.Form.Get(value.(string))
		if len(v) == 0 {
			return "", errors.New(
				fmt.Sprintf(
					"request post parameter \"%s\" can not be found. Please make sure it is set",
					value.(string)))
		}
		return anyToStringValue(v)
	case "http_header":
		v := request.Header.Get(value.(string))
		if len(v) == 0 {
			return "", errors.New(
				fmt.Sprintf(
					"request header parameter \"%s\" can not be found. Please make sure it is set",
					value.(string)))
		}
		return anyToStringValue(v)
	case "env_var":
		v, found := os.LookupEnv(value.(string))
		if !found {
			return "", errors.New(
				fmt.Sprintf(
					"environment variable \"%s\" can not be found. Please make sure it is set",
					value.(string)))
		}
		return anyToStringValue(v)
	case "secret":
		b, err := os.ReadFile("/run/secrets/" + value.(string))
		if err != nil {
			return "", errors.New(
				fmt.Sprintf(
					"secret \"%s\" can not be read. Please make sure it is mounted",
					value.(string)))
		}
		return string(b), nil
	case "file":
		b, err := os.ReadFile(value.(string))
		if err != nil {
			return "", errors.New(
				fmt.Sprintf("file \"%s\" can not be read. Please make sure it is readable", value.(string)))
		}
		return string(b), nil
	case "etcd":
		kv, kvErr := etcdKeyValue(etcdClient, value.(string))
		if kvErr != nil {
			return "", fmt.Errorf("etcd key \"%s\" can not be read", value.(string))
		}

		return kv.String(), nil
	}

	return "", errors.New("failed to retrieve string parameter")
}

func retrieveIntParameterValue(
	valueType string,
	value any,
	request *http.Request,
	etcdClient *clientv3.Client,
) (int64, error) {
	switch valueType {
	case "value":
		return anyToIntValue(value)
	case "env_var":
		envVarName := value.(string)
		env, found := os.LookupEnv(envVarName)
		if !found {
			return 0, errors.New(
				fmt.Sprintf(
					"environment variable \"%s\" can not be found. Please make sure it is set",
					envVarName))
		}
		return anyToIntValue(env)
	case "secret":
		secretName := value.(string)
		b, err := os.ReadFile("/run/secrets/" + secretName)
		if err != nil {
			return 0, errors.New(
				fmt.Sprintf(
					"secret \"%s\" can not be read. Please make sure it is mounted",
					secretName))
		}
		return anyToIntValue(string(b))
	case "file":
		filename := value.(string)
		b, err := os.ReadFile(filename)
		if err != nil {
			return 0, errors.New(
				fmt.Sprintf("file \"%s\" can not be read. Please make sure it is readable", filename))
		}
		return anyToIntValue(string(b))
	case "http_get":
		getParameterName := value.(string)
		v := request.URL.Query().Get(getParameterName)
		if len(v) == 0 {
			return 0, errors.New(
				fmt.Sprintf(
					"request get parameter \"%s\" can not be found. Please make sure it is set",
					getParameterName))
		}
		return anyToIntValue(v)
	case "http_post":
		err := request.ParseForm()
		if err != nil {
			return 0, errors.New("request post parameters can not be parsed")
		}

		postParameterName := value.(string)
		v := request.Form.Get(postParameterName)
		if len(v) == 0 {
			return 0, errors.New(
				fmt.Sprintf(
					"request post parameter \"%s\" can not be found. Please make sure it is set",
					postParameterName))
		}
		return anyToIntValue(v)
	case "http_header":
		headerName := value.(string)
		v := request.Header.Get(headerName)
		if len(v) == 0 {
			return 0, errors.New(
				fmt.Sprintf(
					"request header parameter \"%s\" can not be found. Please make sure it is set",
					headerName))
		}
		return anyToIntValue(v)
	case "etcd":
		kv, kvErr := etcdKeyValue(etcdClient, value.(string))
		if kvErr != nil {
			return 0, fmt.Errorf("etcd key \"%s\" can not be read", value.(string))
		}

		return anyToIntValue(kv.String())
	}

	return 0, errors.New("failed to retrieve integer parameter")
}

func retrieveBoolParameterValue(
	valueType string,
	value any,
	request *http.Request,
	etcdClient *clientv3.Client,
) (bool, error) {
	switch valueType {
	case "value":
		return anyToBoolValue(value)
	case "env_var":
		envVarName := value.(string)
		env, found := os.LookupEnv(envVarName)
		if !found {
			return false, errors.New(
				fmt.Sprintf(
					"environment variable \"%s\" can not be found. Please make sure it is set",
					envVarName))
		}
		return anyToBoolValue(env)
	case "secret":
		secretName := value.(string)
		b, err := os.ReadFile("/run/secrets/" + secretName)
		if err != nil {
			return false, errors.New(
				fmt.Sprintf(
					"secret \"%s\" can not be read. Please make sure it is mounted",
					secretName))
		}
		return anyToBoolValue(string(b))
	case "file":
		filename := value.(string)
		b, err := os.ReadFile(filename)
		if err != nil {
			return false, errors.New(
				fmt.Sprintf("file \"%s\" can not be read. Please make sure it is readable", filename))
		}
		return anyToBoolValue(string(b))
	case "http_get":
		getParameterName := value.(string)
		v := request.URL.Query().Get(getParameterName)
		if len(v) == 0 {
			return false, errors.New(
				fmt.Sprintf(
					"request get parameter \"%s\" can not be found. Please make sure it is set",
					getParameterName))
		}
		return anyToBoolValue(v)
	case "http_post":
		err := request.ParseForm()
		if err != nil {
			return false, errors.New("request post parameters can not be parsed")
		}

		postParameterName := value.(string)
		v := request.Form.Get(postParameterName)
		if len(v) == 0 {
			return false, errors.New(
				fmt.Sprintf(
					"request post parameter \"%s\" can not be found. Please make sure it is set",
					postParameterName))
		}
		return anyToBoolValue(v)
	case "http_header":
		headerName := value.(string)
		v := request.Header.Get(headerName)
		if len(v) == 0 {
			return false, errors.New(
				fmt.Sprintf(
					"request header parameter \"%s\" can not be found. Please make sure it is set",
					headerName))
		}
		return anyToBoolValue(v)
	case "etcd":
		kv, kvErr := etcdKeyValue(etcdClient, value.(string))
		if kvErr != nil {
			return false, fmt.Errorf("etcd key \"%s\" can not be read", value.(string))
		}

		return anyToBoolValue(kv.String())
	}

	return false, errors.New("failed to retrieve integer parameter")
}

func etcdKeyValue(etcdClient *clientv3.Client, key string) (*mvccpb.KeyValue, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	resp, respErr := etcdClient.Get(ctx, key)
	cancel()
	if respErr != nil {
		common.Logger.Warn(
			"etcd get response error",
			slog.String("key", key),
			slog.Any("error", respErr))

		return nil, fmt.Errorf("etcd key \"%s\" can not be read", key)
	}

	if len(resp.Kvs) != 0 {
		return resp.Kvs[0], nil
	}

	return nil, fmt.Errorf("etcd key \"%s\" value can not be found", key)
}

func anyToIntValue(value any) (int64, error) {
	if v, ok := value.(string); ok {
		return strconv.ParseInt(v, 10, 64)
	}

	return int64(value.(float64)), nil
}

func anyToStringValue(value any) (string, error) {
	return value.(string), nil
}

func anyToBoolValue(value any) (bool, error) {
	if reflect.TypeOf(value).Kind() == reflect.String {
		return strconv.ParseBool(value.(string))
	}

	if reflect.TypeOf(value).Kind() == reflect.Int {
		return value.(int) != 0, nil
	}

	return value.(bool), nil
}

func anyToStringArrayValue(value any) []string {
	if v, ok := value.([]interface{}); ok {
		values := make([]string, len(v))
		for i := range v {
			if sv, ok := v[i].(string); ok {
				values[i] = sv
			}
		}

		return values
	}

	if v, ok := value.([]string); ok {
		return v
	}

	return nil
}
