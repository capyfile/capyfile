package operations

import (
	"capyfile/capyerr"
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
	file, err := capyfs.CopyOnWriteFilesystem.Open("testdata/file_5kb.bin")
	if err != nil {
		t.Fatal(err)
	}

	processableFile := files.NewProcessableFile(file)
	in := []files.ProcessableFile{
		*processableFile,
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

				if params.Body != processableFile.File {
					t.Fatalf("expected body to be %v, got %v", processableFile.File, params.Body)
				}

				return &s3.PutObjectOutput{}, nil
			},
		),
	}
	out, err := operation.Handle(in)
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
	file, err := capyfs.CopyOnWriteFilesystem.Open("testdata/file_5kb.bin")
	if err != nil {
		t.Fatal(err)
	}

	processableFile := files.NewProcessableFile(file)
	in := []files.ProcessableFile{
		*processableFile,
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

				if params.Body != processableFile.File {
					t.Fatalf("expected body to be %v, got %v", processableFile.File, params.Body)
				}

				return &s3.PutObjectOutput{}, errors.New("whatever error")
			},
		),
	}
	_, err = operation.Handle(in)

	if err == nil {
		t.Fatal("expected S3 upload error, got nil")
	}

	var ocType *capyerr.OperationConfigurationType
	if errors.As(err, &ocType) {
		if ocType.Code() != ErrorCodeS3UploadOperationConfiguration {
			t.Fatalf(
				"expected error %s code, got error %s code",
				ErrorCodeS3UploadOperationConfiguration,
				ocType.Code(),
			)
		}
	} else {
		t.Fatalf("expected %v error, got %v error", ocType, err)
	}
}
