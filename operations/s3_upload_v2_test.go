package operations

import (
	"capyfile/capyerr"
	"capyfile/capyfs"
	"capyfile/files"
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"testing"
)

type mockPutObjectV2API func(
	ctx context.Context,
	params *s3.PutObjectInput,
	optFns ...func(options *s3.Options),
) (*s3.PutObjectOutput, error)

func (m mockPutObjectV2API) PutObject(
	ctx context.Context,
	params *s3.PutObjectInput,
	optFns ...func(options *s3.Options),
) (*s3.PutObjectOutput, error) {
	return m(ctx, params, optFns...)
}

func TestS3UploadV2Operation_HandleSuccessfulFilesUpload(t *testing.T) {
	capyfs.InitCopyOnWriteFilesystem()

	file, err := capyfs.Filesystem.Open("testdata/file_5kb.bin")
	if err != nil {
		t.Fatal(err)
	}

	processableFile := files.NewProcessableFile(file)
	in := []files.ProcessableFile{
		processableFile,
	}

	operation := &S3UploadV2Operation{
		Params: &S3UploadV2OperationParams{
			Endpoint: "example.com",
			Bucket:   "files",
		},
		PutObjectAPI: mockPutObjectV2API(
			func(
				ctx context.Context,
				params *s3.PutObjectInput,
				optFns ...func(options *s3.Options),
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

	if v, ok := out[0].OperationMetadata[MetadataKeyS3UploadV2FileUrl]; ok {
		expected := "https://files.example.com/" + out[0].GeneratedFilename()
		if v != expected {
			t.Fatalf(
				"metadata key %s = %s, want %s",
				MetadataKeyS3UploadV2FileUrl,
				v,
				expected,
			)
		}
	} else {
		t.Fatalf(
			"metadata key %s = nil, want !nil",
			MetadataKeyS3UploadV2FileUrl,
		)
	}
}

func TestS3UploadV2Operation_HandleFailedFilesUpload(t *testing.T) {
	capyfs.InitCopyOnWriteFilesystem()

	file, err := capyfs.Filesystem.Open("testdata/file_5kb.bin")
	if err != nil {
		t.Fatal(err)
	}

	processableFile := files.NewProcessableFile(file)
	in := []files.ProcessableFile{
		processableFile,
	}

	operation := &S3UploadV2Operation{
		Params: &S3UploadV2OperationParams{
			Endpoint: "example.com",
			Bucket:   "files",
		},
		PutObjectAPI: mockPutObjectV2API(
			func(
				ctx context.Context,
				params *s3.PutObjectInput,
				optFns ...func(options *s3.Options),
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
		if ocType.Code() != ErrorCodeS3UploadV2OperationConfiguration {
			t.Fatalf(
				"expected error %s code, got error %s code",
				ErrorCodeS3UploadV2OperationConfiguration,
				ocType.Code(),
			)
		}
	} else {
		t.Fatalf("expected %v error, got %v error", ocType, err)
	}
}
