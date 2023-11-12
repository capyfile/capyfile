package capysvc

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// So far Service is a service definition. Can be extended in the future.
var serviceDefinition *Service

func FindService(name string) *Service {
	if serviceDefinition.Name == name {
		return serviceDefinition
	}

	return nil
}

func LoadServiceDefinition(serviceDefinitionFile string) error {
	if serviceDefinitionFile != "" {
		serviceDef, err := NewServiceDefinitionFromFile(serviceDefinitionFile)
		if err != nil {
			return err
		}

		serviceDefinition = serviceDef

		return nil
	}

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

	serviceDef, err := NewServiceDefinitionFromFile("/etc/capyfile/service-definition.json")
	if err != nil {
		return err
	}

	serviceDefinition = serviceDef

	return nil
}

func LoadTestServiceDefinition(testServiceDef *Service) {
	serviceDefinition = testServiceDef
}

func NewServiceDefinitionFromFile(filename string) (*Service, error) {
	serviceDefBytes, readErr := os.ReadFile(filename)
	if readErr != nil {
		return nil, readErr
	}

	ext := filepath.Ext(filename)
	if ext == ".yaml" || ext == ".yml" {
		return newServiceDefinitionFromYaml(serviceDefBytes)
	}

	return newServiceDefinitionFromJson(serviceDefBytes)
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

	return serviceDef, err
}

func newServiceDefinitionFromYaml(serviceDefYaml []byte) (serviceDef *Service, err error) {
	err = yaml.Unmarshal(serviceDefYaml, &serviceDef)

	return serviceDef, err
}
