package parameters

type ParameterLoaderProvider interface {
	// HasParameterLoader Whether this parameter loader provider can load parameter from
	// the given source type.
	HasParameterLoader(sourceType string) bool
	// ParameterLoader Returns parameter loader implementation.
	//
	// Source type tells from where to load the parameter value. For example, it can be "value",
	// "env_var", "file", "http_get", "http_header", etc.
	//
	// For example, if the source type is "value" - the source contains the actual value of
	// the parameter. If the source type is "env_var" - the source contains the environment
	// variable name.
	ParameterLoader(sourceType string, source any) (ParameterLoader, error)
}

// ParameterLoader Interface to implement parameter loaders.
type ParameterLoader interface {
	LoadIntValue() (int64, error)
	LoadStringValue() (string, error)
	LoadStringArrayValue() ([]string, error)
}
