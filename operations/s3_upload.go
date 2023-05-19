package operations

import (
	"capyfile/capyerr"
	"capyfile/files"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"net/url"
	"strings"
)

const ErrorCodeS3UploadOperationConfiguration = "S3_UPLOAD_OPERATION_CONFIGURATION"

const MetadataKeyS3UploadFileUrl = "s3_upload.file_url"

// PutObjectAPI The interface to implement PutObjectWithContext that we need to upload the files to S3.
type PutObjectAPI interface {
	PutObjectWithContext(
		ctx context.Context,
		params *s3.PutObjectInput,
		opts ...request.Option,
	) (*s3.PutObjectOutput, error)
}

type S3UploadOperation struct {
	Params       *S3UploadOperationParams
	PutObjectAPI PutObjectAPI
}

type S3UploadOperationParams struct {
	AccessKeyId     string
	SecretAccessKey string
	SessionToken    string
	Endpoint        string
	Region          string
	Bucket          string
}

// compileEndpoint Compiles the endpoint based on the provided pattern.
//
// The pattern can contain the following placeholders:
// - {bucket} - S3 bucket
// - {region} - S3 region
//
// This gives more options to configure the endpoint for different S3 storage providers. For example:
// - s3.amazonaws.com
// - https://{region}.digitaloceanspaces.com/{bucket}
// - https://play.min.io/{bucket}
// - http://minio.local/{bucket}
//
// Another option is to build a custom endpoint resolver for AWS S3 DSK. But this would require a lot
// of additional work.
func (p *S3UploadOperationParams) compileEndpoint() (string, error) {
	replacer := strings.NewReplacer(
		"{bucket}", p.Bucket,
		"{region}", p.Region,
	)
	return replacer.Replace(p.Endpoint), nil
}

// compileFileUrl Compiles a file URL based on the available parameters and provided key.
//
// Compatibility of this solution is not really great. But should work for all major providers.
// Another option is to use the GetObject API to get the file URL. But this would require additional
// permissions and would be slower.
func (p *S3UploadOperationParams) compileFileUrl(key string) (string, error) {
	endpoint, _ := p.compileEndpoint()

	if p.Endpoint == "s3.amazonaws.com" {
		return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", p.Bucket, key), nil
	}

	u, urlParseErr := url.Parse(endpoint)
	if urlParseErr != nil {
		return "", urlParseErr
	}

	u.Path = fmt.Sprintf("%s/%s", u.Path, key)

	return u.String(), nil
}

func (o *S3UploadOperation) Handle(in []files.ProcessableFile) ([]files.ProcessableFile, error) {
	if o.PutObjectAPI == nil {
		initErr := o.InitPutObjectAPI()
		if initErr != nil {
			return in, initErr
		}
	}

	for i := range in {
		var processableFile = &in[i]

		if processableFile.HasFileProcessingError() {
			continue
		}

		// Ensure that we are at the beginning of the file so S3 SDK starts reading it from the beginning.
		_, fsErr := processableFile.File.Seek(0, 0)
		if fsErr != nil {
			return in, fsErr
		}

		_, err := o.PutObjectAPI.PutObjectWithContext(context.TODO(), &s3.PutObjectInput{
			Bucket: aws.String(o.Params.Bucket),
			Key:    aws.String(processableFile.GeneratedFilename()),
			Body:   processableFile.File,
		})
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case s3.ErrCodeNoSuchBucket:
					return in, capyerr.NewOperationConfigurationError(
						ErrorCodeS3UploadOperationConfiguration,
						"storage bucket does not exist",
						err,
					)
				}
			}

			return in, capyerr.NewOperationConfigurationError(
				ErrorCodeS3UploadOperationConfiguration,
				"request to S3 storage has failed",
				err,
			)
		}

		fileUrl, fileUrlError := o.Params.compileFileUrl(processableFile.GeneratedFilename())
		if fileUrlError != nil {
			return in, fileUrlError
		}

		processableFile.AddOperationMetadata(MetadataKeyS3UploadFileUrl, fileUrl)
	}

	return in, nil
}

// InitPutObjectAPI Init PutObjectAPI that we need to upload the files to S3.
func (o *S3UploadOperation) InitPutObjectAPI() error {
	endpoint, endpointErr := o.Params.compileEndpoint()
	if endpointErr != nil {
		return endpointErr
	}

	sess := session.Must(
		session.NewSession(
			&aws.Config{
				Credentials: credentials.NewStaticCredentials(
					o.Params.AccessKeyId,
					o.Params.SecretAccessKey,
					o.Params.SessionToken,
				),
				Endpoint: &endpoint,
				Region:   &o.Params.Region,
			},
		),
	)
	o.PutObjectAPI = s3.New(sess)

	return nil
}
