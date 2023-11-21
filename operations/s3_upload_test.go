package operations

import (
	"bytes"
	"capyfile/capyfs"
	"capyfile/files"
	"context"
	"errors"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"testing"
)

type mockPutObjectAPI func(
	ctx context.Context,
	params *s3.PutObjectInput,
	opts ...request.Option,
) (*s3.PutObjectOutput, error)

func (m mockPutObjectAPI) PutObjectWithContext(
	ctx context.Context,
	params *s3.PutObjectInput,
	opts ...request.Option,
) (*s3.PutObjectOutput, error) {
	return m(ctx, params, opts...)
}

func TestS3UploadOperation_HandleSuccessfulFilesUpload(t *testing.T) {
	capyfs.InitCopyOnWriteFilesystem()

	file, err := capyfs.Filesystem.Open("testdata/file_5kb.bin")
	if err != nil {
		t.Fatal(err)
	}

	processableFile := files.NewProcessableFile(file.Name())
	in := []files.ProcessableFile{
		processableFile,
	}

	operation := &S3UploadOperation{
		Params: &S3UploadOperationParams{
			Endpoint: "https://{bucket}.example.com",
			Bucket:   "files",
		},
		PutObjectAPI: mockPutObjectAPI(
			func(
				ctx context.Context,
				params *s3.PutObjectInput,
				opts ...request.Option,
			) (*s3.PutObjectOutput, error) {
				t.Helper()

				if *params.Bucket != "files" {
					t.Fatalf("expected bucket to be files, got %v", params.Bucket)
				}

				if *params.Key != processableFile.GeneratedFilename() {
					t.Fatalf("expected key to be %s, got %v", processableFile.GeneratedFilename(), params.Bucket)
				}

				bodyContent := make([]byte, 5120)
				_, readErr := params.Body.Read(bodyContent)
				if readErr != nil {
					t.Fatal(readErr)
				}

				fileContent := make([]byte, 5120)
				_, readErr = file.Read(fileContent)
				if readErr != nil {
					t.Fatal(readErr)
				}

				if bytes.Compare(bodyContent, fileContent) != 0 {
					t.Fatalf("expected body content to be equal to file content")
				}

				return &s3.PutObjectOutput{}, nil
			},
		),
	}
	out, err := operation.Handle(in, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(out) != 1 {
		t.Fatalf("len(out) = %d, want 1", len(out))
	}

	if out[0].FileProcessingError != nil {
		t.Fatalf(
			"FileProcessingError.Code() = %s, want nil",
			out[0].FileProcessingError.Code(),
		)
	}

	if v, ok := out[0].OperationMetadata[MetadataKeyS3UploadFileUrl]; ok {
		expected := "https://files.example.com/" + out[0].GeneratedFilename()
		if v != expected {
			t.Fatalf(
				"metadata key %s = %s, want %s",
				MetadataKeyS3UploadFileUrl,
				v,
				expected,
			)
		}
	} else {
		t.Fatalf(
			"metadata key %s = nil, want !nil",
			MetadataKeyS3UploadFileUrl,
		)
	}
}

func TestS3UploadOperation_HandleFailedFilesUpload(t *testing.T) {
	capyfs.InitCopyOnWriteFilesystem()

	file, err := capyfs.Filesystem.Open("testdata/file_5kb.bin")
	if err != nil {
		t.Fatal(err)
	}

	processableFile := files.NewProcessableFile(file.Name())
	in := []files.ProcessableFile{
		processableFile,
	}

	operation := &S3UploadOperation{
		Params: &S3UploadOperationParams{
			Endpoint: "example.com",
			Bucket:   "files",
		},
		PutObjectAPI: mockPutObjectAPI(
			func(
				ctx context.Context,
				params *s3.PutObjectInput,
				opts ...request.Option,
			) (*s3.PutObjectOutput, error) {
				t.Helper()

				if *params.Bucket != "files" {
					t.Fatalf("expected bucket to be files, got %v", params.Bucket)
				}

				if *params.Key != processableFile.GeneratedFilename() {
					t.Fatalf("expected key to be %s, got %v", processableFile.GeneratedFilename(), params.Bucket)
				}

				bodyContent := make([]byte, 5120)
				_, readErr := params.Body.Read(bodyContent)
				if readErr != nil {
					t.Fatal(readErr)
				}

				fileContent := make([]byte, 5120)
				_, readErr = file.Read(fileContent)
				if readErr != nil {
					t.Fatal(readErr)
				}

				if bytes.Compare(bodyContent, fileContent) != 0 {
					t.Fatalf("expected body content to be equal to file content")
				}

				return &s3.PutObjectOutput{}, errors.New("whatever error")
			},
		),
	}

	out, handleErr := operation.Handle(in, nil, nil)

	if handleErr != nil {
		t.Fatal(handleErr)
	}

	if len(out) != 1 {
		t.Fatalf("len(out) = %d, want 1", len(out))
	}

	if out[0].FileProcessingError == nil {
		t.Fatalf("FileProcessingError = nil, want !nil")
	}

	var s3FileUploadFailureError *S3FileUploadFailureError
	errors.As(out[0].FileProcessingError, &s3FileUploadFailureError)

	if s3FileUploadFailureError == nil {
		t.Fatalf("FileProcessingError = nil, want !nil")
	}

	if s3FileUploadFailureError.Code() != ErrorCodeS3FileUploadFailure {
		t.Fatalf(
			"FileProcessingError.Code() = %s, want %s",
			s3FileUploadFailureError.Code(),
			ErrorCodeS3FileUploadFailure,
		)
	}

	if s3FileUploadFailureError.Data.OrigErr.Error() != "whatever error" {
		t.Fatalf(
			"FileProcessingError.Error() = %s, want %s",
			out[0].FileProcessingError.Error(),
			"whatever error",
		)
	}
}
