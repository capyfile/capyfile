package operations

import (
	"capyfile/capyerr"
	"capyfile/files"
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const ErrorCodeS3UploadV2OperationConfiguration = "S3_UPLOAD_OPERATION_CONFIGURATION"

const MetadataKeyS3UploadV2FileUrl = "s3_upload.file_url"

// PutObjectV2API The interface to implement PutObject that we need to upload the files to S3.
type PutObjectV2API interface {
	PutObject(
		ctx context.Context,
		params *s3.PutObjectInput,
		optFns ...func(options *s3.Options),
	) (*s3.PutObjectOutput, error)
}

type S3UploadV2Operation struct {
	Params       *S3UploadV2OperationParams
	PutObjectAPI PutObjectV2API
}

type S3UploadV2OperationParams struct {
	AccessKeyId     string
	SecretAccessKey string
	SessionToken    string
	Endpoint        string
	Region          string
	Bucket          string
}

func (o *S3UploadV2Operation) Handle(in []files.ProcessableFile) ([]files.ProcessableFile, error) {
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
		_, fileSeekErr := processableFile.File.Seek(0, 0)
		if fileSeekErr != nil {
			return in, fileSeekErr
		}

		_, putObjectErr := o.PutObjectAPI.PutObject(context.TODO(), &s3.PutObjectInput{
			Bucket: aws.String(o.Params.Bucket),
			Key:    aws.String(processableFile.GeneratedFilename()),
			Body:   processableFile.File,
		})
		if putObjectErr != nil {
			var nsb *types.NoSuchBucket
			if errors.As(putObjectErr, &nsb) {
				return in, capyerr.NewOperationConfigurationError(
					ErrorCodeS3UploadV2OperationConfiguration,
					"storage bucket does not exist",
					putObjectErr,
				)
			}

			return in, capyerr.NewOperationConfigurationError(
				ErrorCodeS3UploadV2OperationConfiguration,
				"request to S3 storage has failed",
				putObjectErr,
			)
		}

		// Works for AWS S3, but not sure about the compatibility with the other S3 storage providers here.
		processableFile.AddOperationMetadata(
			MetadataKeyS3UploadV2FileUrl,
			fmt.Sprintf("https://%s.%s/%s", o.Params.Bucket, o.Params.Endpoint, processableFile.GeneratedFilename()),
		)
	}

	return in, nil
}

// InitPutObjectAPI Init PutObjectV2API that we need to upload the files to S3.
func (o *S3UploadV2Operation) InitPutObjectAPI() error {
	cfg, cfgErr := config.LoadDefaultConfig(
		context.TODO(),
		config.WithCredentialsProvider(
			aws.CredentialsProviderFunc(func(context context.Context) (aws.Credentials, error) {
				return aws.Credentials{
					AccessKeyID:     o.Params.AccessKeyId,
					SecretAccessKey: o.Params.SecretAccessKey,
					SessionToken:    o.Params.SessionToken,
				}, nil
			}),
		),
		config.WithRegion(o.Params.Region),
		// todo: it does not work like this, need more work with the endpoint resolver
		//config.WithEndpointResolverWithOptions(
		//	aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		//		if service == s3.ServiceID {
		//			return aws.Endpoint{
		//				URL: o.Params.Endpoint,
		//			}, nil
		//		}
		//		// returning EndpointNotFoundError will allow the service to fallback to it's default resolution
		//		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
		//	}),
		//),
	)
	if cfgErr != nil {
		return cfgErr
	}

	o.PutObjectAPI = s3.NewFromConfig(cfg)

	return nil
}
