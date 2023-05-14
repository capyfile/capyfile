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

		// Works for AWS S3, but not sure about the compatibility with the other S3 storage providers here.
		processableFile.AddOperationMetadata(
			MetadataKeyS3UploadFileUrl,
			fmt.Sprintf("https://%s.%s/%s", o.Params.Bucket, o.Params.Endpoint, processableFile.GeneratedFilename()),
		)
	}

	return in, nil
}

// InitPutObjectAPI Init PutObjectAPI that we need to upload the files to S3.
func (o *S3UploadOperation) InitPutObjectAPI() error {
	sess := session.Must(
		session.NewSession(
			&aws.Config{
				Credentials: credentials.NewStaticCredentials(
					o.Params.AccessKeyId,
					o.Params.SecretAccessKey,
					o.Params.SessionToken,
				),
				Endpoint: &o.Params.Endpoint,
				Region:   &o.Params.Region,
			},
		),
	)
	o.PutObjectAPI = s3.New(sess)

	return nil
}
