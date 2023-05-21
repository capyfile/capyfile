package capysvc

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
)

// So far Service is a service definition. Can be extended in the future.
var serviceDefinition *Service

func FindService(name string) *Service {
	if serviceDefinition.Name == name {
		return serviceDefinition
	}

	return nil
}

func LoadServiceDefinition() error {
	sdUrl, sdUrlFound := os.LookupEnv("CAPYFILE_SERVICE_DEFINITION_URL")
	if sdUrlFound {
		serviceDef, err := NewServiceDefinitionFromUrl(sdUrl)
		if err != nil {
			return err
		}

		serviceDefinition = serviceDef

		return nil
	}

	sdFilename, sdFileFound := os.LookupEnv("CAPYFILE_SERVICE_DEFINITION_FILE")
	if sdFileFound {
		serviceDef, err := NewServiceDefinitionFromFile(sdFilename)
		if err != nil {
			return err
		}

		serviceDefinition = serviceDef

		return nil
	}

	return errors.New("service definition source is not provided")
}

func LoadTestServiceDefinition(testServiceDef *Service) {
	serviceDefinition = testServiceDef
}

func NewServiceDefinitionFromFile(filename string) (*Service, error) {
	serviceDefJson, readErr := os.ReadFile(filename)
	if readErr != nil {
		return nil, readErr
	}

	return newServiceDefinitionFromJson(serviceDefJson)
}

func NewServiceDefinitionFromUrl(url string) (*Service, error) {
	res, resErr := http.Get(url)
	if resErr != nil {
		return nil, resErr
	}

	serviceDefJson, readErr := io.ReadAll(res.Body)
	_ = res.Body.Close()
	if readErr != nil {
		return nil, readErr
	}

	return newServiceDefinitionFromJson(serviceDefJson)
}

func newServiceDefinitionFromJson(serviceDefJson []byte) (serviceDef *Service, err error) {
	err = json.Unmarshal(serviceDefJson, &serviceDef)
	if err != nil {
		return serviceDef, err
	}

	return serviceDef, nil
}
