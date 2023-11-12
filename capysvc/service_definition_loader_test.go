package capysvc

import (
	"golang.org/x/exp/slices"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
)

func TestLoadServiceDefinitionFromJsonFile(t *testing.T) {
	setenvErr := os.Setenv("CAPYFILE_SERVICE_DEFINITION_FILE", "testdata/service-definition.json")
	if setenvErr != nil {
		t.Fatalf("expected no error while setting CAPYFILE_SERVICE_DEFINITION_FILE env varm got %v", setenvErr)
	}

	sdErr := LoadServiceDefinition("")
	if sdErr != nil {
		t.Fatalf("expected no error while loading service definition, got %v", sdErr)
	}

	if serviceDefinition == nil {
		t.Fatalf("expected loaded ServiceDevinition, got nil")
	}

	serviceDefinitionAssertions(serviceDefinition, t)
}

func TestLoadServiceDefinitionFromYamlFile(t *testing.T) {
	setenvErr := os.Setenv("CAPYFILE_SERVICE_DEFINITION_FILE", "testdata/service-definition.yml")
	if setenvErr != nil {
		t.Fatalf("expected no error while setting CAPYFILE_SERVICE_DEFINITION_FILE env varm got %v", setenvErr)
	}

	sdErr := LoadServiceDefinition("")
	if sdErr != nil {
		t.Fatalf("expected no error while loading service definition, got %v", sdErr)
	}

	if serviceDefinition == nil {
		t.Fatalf("expected loaded ServiceDevinition, got nil")
	}

	serviceDefinitionAssertions(serviceDefinition, t)
}

func TestLoadServiceDefinitionFromURL(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		sdJsonFile, sdFileReadErr := os.Open("testdata/service-definition.json")
		if sdFileReadErr != nil {
			t.Fatalf(
				"expect no error while reading service definition test data file, got error %v",
				sdFileReadErr)
		}
		defer sdJsonFile.Close()

		body, sdReadErr := io.ReadAll(sdJsonFile)
		if sdReadErr != nil {
			t.Fatalf(
				"expect no error while reading service definition test data file content, got error %v",
				sdReadErr)
		}

		_, sdWriteErr := w.Write(body)
		if sdWriteErr != nil {
			t.Fatalf(
				"expect no error while writing service definition test data, got error %v",
				sdWriteErr)
		}
	}))
	defer testServer.Close()

	setenvErr := os.Setenv("CAPYFILE_SERVICE_DEFINITION_URL", testServer.URL)
	if setenvErr != nil {
		t.Fatalf(
			"expected no error while setting CAPYFILE_SERVICE_DEFINITION_URL env varm got %v",
			setenvErr)
	}

	sdErr := LoadServiceDefinition("")
	if sdErr != nil {
		t.Fatalf("expected no error while loading service definition, got %v", sdErr)
	}

	if serviceDefinition == nil {
		t.Fatalf("expected loaded ServiceDevinition, got nil")
	}

	serviceDefinitionAssertions(serviceDefinition, t)
}

type operationTestCase struct {
	name   string
	params []operationParamTestCase
}
type operationParamTestCase struct {
	paramName  string
	sourceType string
	source     any
}

func serviceDefinitionAssertions(sd *Service, t *testing.T) {
	if sd.Version != "1.1" {
		t.Fatalf("serviceDefinition.Version = %s, want 1.0", sd.Version)
	}

	if sd.Name != "avatars" {
		t.Fatalf("serviceDefinition.Name = %s, want messenger", sd.Name)
	}

	if len(sd.Processors) != 1 {
		t.Fatalf("len(serviceDefinition.Processors) = %d, want 1", len(sd.Processors))
	}

	idx := slices.IndexFunc(sd.Processors, func(p Processor) bool {
		return p.Name == "upload"
	})
	if idx == -1 {
		t.Fatalf("FindProcessor.Name != avatar, want FindProcessor.Name == avatar")
	}

	for _, p := range sd.Processors {
		if p.Name == "upload" {
			if len(p.Operations) != 5 {
				t.Fatalf("len(FindProcessor.Operations) = %d, want 4", len(p.Operations))
			}

			operationNamesCases := []string{
				"http_multipart_form_data_input_read",
				"file_size_validate",
				"file_type_validate",
				"metadata_cleanup",
				"s3_upload",
			}
			for _, c := range operationNamesCases {
				idx = slices.IndexFunc(p.Operations, func(o Operation) bool {
					return o.Name == c
				})
				if idx == -1 {
					t.Fatalf("Operation.Name != %s, want Operation.Name == %s", c, c)
				}
			}

			operationParamsCases := []operationTestCase{
				{
					name:   "http_multipart_form_data_input_read",
					params: []operationParamTestCase{},
				},
				{
					name: "file_size_validate",
					params: []operationParamTestCase{
						{
							paramName:  "maxFileSize",
							sourceType: "value",
							source:     float64(1048576),
						},
					},
				},
				{
					name: "file_type_validate",
					params: []operationParamTestCase{
						{
							paramName:  "allowedMimeTypes",
							sourceType: "value",
							source:     []interface{}{"image/jpeg", "image/png", "image/heif"},
						},
					},
				},
				{
					name:   "metadata_cleanup",
					params: []operationParamTestCase{},
				},
				{
					name: "s3_upload",
					params: []operationParamTestCase{
						{
							paramName:  "accessKeyId",
							sourceType: "env_var",
							source:     "AWS_ACCESS_KEY_ID",
						},
						{
							paramName:  "secretAccessKey",
							sourceType: "env_var",
							source:     "AWS_SECRET_ACCESS_KEY",
						},
						{
							paramName:  "sessionToken",
							sourceType: "env_var",
							source:     "AWS_SESSION_TOKEN",
						},
						{
							paramName:  "endpoint",
							sourceType: "env_var",
							source:     "AWS_ENDPOINT",
						},
						{
							paramName:  "region",
							sourceType: "env_var",
							source:     "AWS_REGION",
						},
						{
							paramName:  "bucket",
							sourceType: "env_var",
							source:     "AWS_AVATAR_BUCKET",
						},
					},
				},
			}

			for _, o := range p.Operations {
				for _, c := range operationParamsCases {
					if o.Name == c.name {
						for _, p := range c.params {
							if v, ok := o.Params[p.paramName]; ok {
								if v.SourceType != p.sourceType {
									t.Fatalf(
										"Opeartion[%s].%s.source = %s, want %s",
										o.Name, p.paramName, v.Source, p.source)
								}
								// json parser return float64 for numbers
								// yaml parser return int64 for numbers
								// We need to make them somehow comparable
								if p.paramName == "maxFileSize" {
									var actualValue int64 = 0
									switch reflect.ValueOf(v.Source).Kind() {
									case reflect.Float64:
										actualValue = int64(reflect.ValueOf(v.Source).Float())
									case reflect.Int64, reflect.Int:
										actualValue = reflect.ValueOf(v.Source).Int()
									default:
										t.Fatalf("Failed to parse Opeartion[%s].%s.source", o.Name, p.paramName)
									}

									var expectedValue int64 = 0
									switch reflect.ValueOf(p.source).Kind() {
									case reflect.Float64:
										expectedValue = int64(reflect.ValueOf(p.source).Float())
									case reflect.Int64, reflect.Int:
										expectedValue = reflect.ValueOf(p.source).Int()
									default:
										t.Fatalf("Failed to parse Opeartion[%s].%s.source", o.Name, p.paramName)
									}

									if expectedValue != actualValue {
										t.Fatalf(
											"Opeartion[%s].%s.source = %v, want %v",
											o.Name, p.paramName, v.Source, p.source)
									}
								} else {
									if !reflect.DeepEqual(v.Source, p.source) {
										t.Fatalf(
											"Opeartion[%s].%s.source = %v, want %v",
											o.Name, p.paramName, v.Source, p.source)
									}
								}
							} else {
								t.Fatalf("Opeartion[%s].%s param not parsed", o.Name, p.paramName)
							}
						}
					}
				}
			}
		}
	}
}
