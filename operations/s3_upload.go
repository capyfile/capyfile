package operations

import (
	"capyfile/capyerr"
	"capyfile/capyfs"
	"capyfile/files"
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/spf13/afero"
	"net/url"
	"strings"
	"sync"
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
	Name         string
	Params       *S3UploadOperationParams
	PutObjectAPI PutObjectAPI
}

func (o *S3UploadOperation) OperationName() string {
	return o.Name
}

func (o *S3UploadOperation) AllowConcurrency() bool {
	return true
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

func (o *S3UploadOperation) Handle(
	in []files.ProcessableFile,
	errorCh chan<- OperationError,
	notificationCh chan<- OperationNotification,
) (out []files.ProcessableFile, err error) {
	if o.PutObjectAPI == nil {
		initErr := o.InitPutObjectAPI()
		if initErr != nil {
			if errorCh != nil {
				errorCh <- o.errorBuilder().Error(
					errors.New("put object API can not be initialized"),
				)
			}

			return out, capyerr.NewOperationConfigurationError(
				ErrorCodeS3UploadOperationConfiguration,
				"put object API can not be initialized",
				initErr,
			)
		}
	}

	var wg sync.WaitGroup

	outHolder := newOutputHolder()

	for i := range in {
		wg.Add(1)

		var pf = &in[i]

		go func(pf *files.ProcessableFile) {
			defer wg.Done()

			if notificationCh != nil {
				notificationCh <- o.notificationBuilder().Started("S3 file upload has started", pf)
			}

			file, fileOpenErr := capyfs.Filesystem.Open(pf.Name())
			if fileOpenErr != nil {
				pf.SetFileProcessingError(
					NewFileCanNotBeOpenedError(fileOpenErr),
				)

				if errorCh != nil {
					errorCh <- o.errorBuilder().ProcessableFileError(pf, fileOpenErr)
				}
				if notificationCh != nil {
					notificationCh <- o.notificationBuilder().Failed(
						"file can not be opened", pf, fileOpenErr)
				}

				outHolder.AppendToOut(pf)

				return
			}
			defer func(file afero.File) {
				closeErr := file.Close()
				if closeErr != nil {
					if errorCh != nil {
						errorCh <- o.errorBuilder().ProcessableFileError(pf, fileOpenErr)
					}
					if notificationCh != nil {
						notificationCh <- o.notificationBuilder().Failed(
							"file can not be opened", pf, fileOpenErr)
					}
				}
			}(file)

			_, putObjErr := o.PutObjectAPI.PutObjectWithContext(context.TODO(), &s3.PutObjectInput{
				Bucket: aws.String(o.Params.Bucket),
				Key:    aws.String(pf.GeneratedFilename()),
				Body:   file,
			})
			if putObjErr != nil {
				pf.SetFileProcessingError(
					NewS3FileUploadFailureError(putObjErr),
				)

				if errorCh != nil {
					errorCh <- o.errorBuilder().ProcessableFileError(pf, putObjErr)
				}

				var aerr awserr.Error
				if errors.As(putObjErr, &aerr) {
					switch aerr.Code() {
					case s3.ErrCodeNoSuchBucket:
						if notificationCh != nil {
							notificationCh <- o.notificationBuilder().Failed(
								"can not upload the file because S3 storage bucket does not exist", pf, putObjErr)
						}
					default:
						if errorCh != nil {
							errorCh <- o.errorBuilder().ProcessableFileError(pf, putObjErr)
						}
						if notificationCh != nil {
							notificationCh <- o.notificationBuilder().Failed(
								"can not upload the file because the request to S3 storage has failed", pf, putObjErr)
						}
					}
				}

				if notificationCh != nil {
					notificationCh <- o.notificationBuilder().Finished("S3 file upload has finished", pf)
				}

				outHolder.AppendToOut(pf)

				return
			}

			fileUrl, fileUrlError := o.Params.compileFileUrl(pf.GeneratedFilename())
			if fileUrlError != nil {
				pf.SetFileProcessingError(
					NewS3FileUrlCanNotBeRetrievedError(fileUrlError),
				)

				if errorCh != nil {
					errorCh <- o.errorBuilder().ProcessableFileError(pf, fileUrlError)
				}
				if notificationCh != nil {
					notificationCh <- o.notificationBuilder().Failed(
						"can not retrieve S3 file URL", pf, fileUrlError)
				}

				outHolder.AppendToOut(pf)

				return
			}

			pf.AddOperationMetadata(MetadataKeyS3UploadFileUrl, fileUrl)

			if notificationCh != nil {
				notificationCh <- o.notificationBuilder().Finished("S3 file upload has finished", pf)
			}

			outHolder.AppendToOut(pf)
		}(pf)
	}

	wg.Wait()

	return outHolder.Out, nil
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

func (o *S3UploadOperation) notificationBuilder() *OperationNotificationBuilder {
	return &OperationNotificationBuilder{
		OperationName: o.Name,
	}
}

func (o *S3UploadOperation) errorBuilder() *OperationErrorBuilder {
	return &OperationErrorBuilder{
		OperationName: o.Name,
	}
}
