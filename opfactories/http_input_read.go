package opfactories

import (
	"capyfile/operations"
	"net/http"
)

func NewHttpMultipartFormInputReadOperation(
	name string,
	req *http.Request,
) (*operations.HttpMultipartFormInputReadOperation, error) {
	return &operations.HttpMultipartFormInputReadOperation{
		Name: name,
		Req:  req,
	}, nil
}

func NewHttpOctetStreamInputReadOperation(
	name string,
	req *http.Request,
) (*operations.HttpOctetStreamInputReadOperation, error) {
	return &operations.HttpOctetStreamInputReadOperation{
		Name: name,
		Req:  req,
	}, nil
}
