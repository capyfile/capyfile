package operations

import (
	"capyfile/capyfs"
	"capyfile/files"
	"fmt"
	"os"
	"testing"
)

type mockCommandExecutor func(name string, arg ...string) (output []byte, err error)

func (m mockCommandExecutor) Execute(name string, arg ...string) (output []byte, err error) {
	return m(name, arg...)
}

func TestCommandExecOperation_Handle(t *testing.T) {
	capyfs.InitCopyOnWriteFilesystem()

	video1Err := capyfs.FilesystemUtils.WriteFile(
		"/tmp/video1.avi", []byte("whatever bytes doesn't matter"), os.ModePerm)
	if video1Err != nil {
		t.Fatal(video1Err)
	}

	video1File, videoFile1Err := capyfs.Filesystem.Open("/tmp/video1.avi")
	if videoFile1Err != nil {
		t.Fatal(videoFile1Err)
	}

	in := []files.ProcessableFile{
		files.NewProcessableFile(video1File.Name()),
	}

	operation := &CommandExecOperation{
		Name: "command_exec",
		Params: &CommandExecOperationParams{
			CommandName: "ffmpeg",
			CommandArgs: []string{
				"-i", "{{.AbsolutePath}}",
				"-c:v", "copy",
				"-c:a", "copy",
				"/tmp/{{.Basename}}.mp4",
			},
			OutputFileDestination: "/tmp/{{.Basename}}.mp4",
		},
		CommandExecutor: mockCommandExecutor(
			func(name string, arg ...string) (output []byte, err error) {
				if name != "ffmpeg" {
					t.Fatalf("expected name to be ffmpeg, got %s", name)
				}

				if len(arg) != 7 {
					t.Fatalf("len(arg) = %d, want 7", len(arg))
				}

				if arg[0] != "-i" {
					t.Fatalf("expected arg[0] to be -i, got %s", arg[0])
				}
				absPath, _ := in[0].FileAbsolutePath()
				if arg[1] != absPath {
					t.Fatalf("expected arg[1] to be %s, got %s", absPath, arg[1])
				}
				if arg[2] != "-c:v" {
					t.Fatalf("expected arg[2] to be -c:v, got %s", arg[2])
				}
				if arg[3] != "copy" {
					t.Fatalf("expected arg[3] to be copy, got %s", arg[3])
				}
				if arg[4] != "-c:a" {
					t.Fatalf("expected arg[4] to be -c:a, got %s", arg[4])
				}
				if arg[5] != "copy" {
					t.Fatalf("expected arg[5] to be copy, got %s", arg[5])
				}
				if arg[6] != fmt.Sprintf("/tmp/%s.mp4", in[0].FileBasename()) {
					t.Fatalf("expected arg[6] to be /tmp/%s.mp4, got %s", in[0].FileBasename(), arg[6])
				}

				// Here the command kind of creates the output file that the operation will
				// parse template for and create a new ProcessableFile for it.
				writeFileErr := capyfs.FilesystemUtils.WriteFile(
					fmt.Sprintf("/tmp/%s.mp4", in[0].FileBasename()),
					[]byte("whatever bytes doesn't matter"),
					os.ModePerm,
				)
				if writeFileErr != nil {
					t.Fatal(writeFileErr)
				}

				return []byte("some cmd output"), nil
			},
		),
	}
	out, opErr := operation.Handle(in, nil, nil)
	if opErr != nil {
		t.Fatal(opErr)
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

	// In addition, we need to check that the new processable file is the output file
	// that the command produced.
	absPath, _ := out[0].FileAbsolutePath()
	if absPath != fmt.Sprintf("/tmp/%s.mp4", in[0].FileBasename()) {
		t.Fatalf(
			"expected file absolute path to be /tmp/%s.mp4, got %s",
			in[0].FileBasename(),
			absPath,
		)
	}
}

func TestCommandExecOperation_HandleEmptyOutput(t *testing.T) {
	capyfs.InitCopyOnWriteFilesystem()

	video1Err := capyfs.FilesystemUtils.WriteFile(
		"/tmp/video1.avi", []byte("whatever bytes doesn't matter"), os.ModePerm)
	if video1Err != nil {
		t.Fatal(video1Err)
	}

	video1File, videoFile1Err := capyfs.Filesystem.Open("/tmp/video1.avi")
	if videoFile1Err != nil {
		t.Fatal(videoFile1Err)
	}

	in := []files.ProcessableFile{
		files.NewProcessableFile(video1File.Name()),
	}

	operation := &CommandExecOperation{
		Name: "command_exec",
		Params: &CommandExecOperationParams{
			CommandName: "aws",
			CommandArgs: []string{
				"s3",
				"cp", "{{.AbsolutePath}}",
				"s3://my-bucket",
				"--region", "us-east-1",
			},
		},
		CommandExecutor: mockCommandExecutor(
			func(name string, arg ...string) (output []byte, err error) {
				if name != "aws" {
					t.Fatalf("expected name to be aws, got %s", name)
				}

				if len(arg) != 6 {
					t.Fatalf("len(arg) = %d, want 6", len(arg))
				}

				if arg[0] != "s3" {
					t.Fatalf("expected arg[0] to be s3, got %s", arg[0])
				}
				if arg[1] != "cp" {
					t.Fatalf("expected arg[1] to be cp, got %s", arg[1])
				}
				absPath, _ := in[0].FileAbsolutePath()
				if arg[2] != absPath {
					t.Fatalf("expected arg[2] to be %s, got %s", absPath, arg[2])
				}
				if arg[3] != "s3://my-bucket" {
					t.Fatalf("expected arg[3] to be s3://my-bucket, got %s", arg[3])
				}
				if arg[4] != "--region" {
					t.Fatalf("expected arg[4] to be --region, got %s", arg[4])
				}
				if arg[5] != "us-east-1" {
					t.Fatalf("expected arg[5] to be us-east-1, got %s", arg[5])
				}

				return []byte("some cmd output"), nil
			},
		),
	}
	out, opErr := operation.Handle(in, nil, nil)
	if opErr != nil {
		t.Fatal(opErr)
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

	// The command has input but no output, so the output file should be the same as
	// the input file.
	outAbsPath, _ := out[0].FileAbsolutePath()
	inAbsPath, _ := in[0].FileAbsolutePath()
	if outAbsPath != inAbsPath {
		t.Fatalf(
			"expected file absolute path to be %s, got %s",
			inAbsPath,
			outAbsPath,
		)
	}
}

func TestCommandExecOperation_HandleEmptyInput(t *testing.T) {
	capyfs.InitCopyOnWriteFilesystem()

	operation := &CommandExecOperation{
		Name: "command_exec",
		Params: &CommandExecOperationParams{
			CommandName: "wget",
			CommandArgs: []string{
				"-P", "/tmp",
				"https://example.com/archive.zip",
			},
			OutputFileDestination: "/tmp/archive.zip",
		},
		CommandExecutor: mockCommandExecutor(
			func(name string, arg ...string) (output []byte, err error) {
				if name != "wget" {
					t.Fatalf("expected name to be wget, got %s", name)
				}

				if len(arg) != 3 {
					t.Fatalf("len(arg) = %d, want 3", len(arg))
				}

				if arg[0] != "-P" {
					t.Fatalf("expected arg[0] to be -P, got %s", arg[0])
				}
				if arg[1] != "/tmp" {
					t.Fatalf("expected arg[1] to be /tmp, got %s", arg[1])
				}
				if arg[2] != "https://example.com/archive.zip" {
					t.Fatalf("expected arg[3] to be https://example.com/archive.zip, got %s", arg[2])
				}

				// Here the command kind of creates the output file that the operation will
				// parse template for and create a new ProcessableFile for it.
				writeFileErr := capyfs.FilesystemUtils.WriteFile(
					"/tmp/archive.zip",
					[]byte("whatever bytes doesn't matter"),
					os.ModePerm,
				)
				if writeFileErr != nil {
					t.Fatal(writeFileErr)
				}

				return []byte("some cmd output"), nil
			},
		),
	}
	out, opErr := operation.Handle([]files.ProcessableFile{}, nil, nil)
	if opErr != nil {
		t.Fatal(opErr)
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

	// In addition, we need to check that the new processable file is the output file
	// that the command produced.
	outAbsPath, _ := out[0].FileAbsolutePath()
	if outAbsPath != "/tmp/archive.zip" {
		t.Fatalf(
			"expected file absolute path to be /tmp/archive.zip, got %s",
			outAbsPath,
		)
	}
}
