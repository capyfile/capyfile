package files

import (
	"capyfile/capyfs"
	"github.com/gabriel-vasile/mimetype"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/spf13/afero"
)

// ProcessableFile The file that can be processed by the operations.
type ProcessableFile struct {
	NanoID string

	File                afero.File
	FileProcessingError FileProcessingError

	Metadata *ProcessableFileMetadata
	// OperationMetadata The metadata related to a specific operation.
	OperationMetadata map[string]interface{}

	mime *mimetype.MIME

	// PreserveOriginalFile Whether the original file should be stored along with the processed one.
	PreserveOriginalFile bool
	// OriginalProcessableFile Sometimes we might need to preserve the original unmodified file
	// to store it along with the modified (processed file, like resized image) file.
	OriginalProcessableFile *ProcessableFile
}

func NewProcessableFile(file afero.File) *ProcessableFile {
	return &ProcessableFile{
		NanoID: gonanoid.Must(),
		File:   file,
	}
}

func NewUnreadableProcessableFile(filepath string, readErr FileProcessingError) *ProcessableFile {
	return &ProcessableFile{
		NanoID: gonanoid.Must(),
		Metadata: &ProcessableFileMetadata{
			OriginalFilename: filepath,
		},
		FileProcessingError: readErr,
	}
}

// ReplaceFile Replaces the file associated with the processable file.
// Here it also updates everything that is related to it, the things like MIME type.
func (f *ProcessableFile) ReplaceFile(file afero.File) {
	if f.PreserveOriginalFile {
		if f.OriginalProcessableFile != nil {
			_ = f.FreeResources()
		} else {
			// If we want to preserve original file and there are no original file associated
			// with this instance, we can consider this instance as the original file.
			f.OriginalProcessableFile = &ProcessableFile{
				File: f.File,
				mime: f.mime,
			}
		}
	} else {
		_ = f.FreeResources()
	}

	f.File = file
	f.mime = nil
}

func (f *ProcessableFile) Mime() (*mimetype.MIME, error) {
	err := f.loadMime()
	if err != nil {
		return nil, err
	}

	return f.mime, nil
}

func (f *ProcessableFile) loadMime() error {
	if f.mime != nil {
		return nil
	}

	stat, err := f.File.Stat()
	if err != nil {
		return err
	}

	b := make([]byte, stat.Size())
	_, err = f.File.ReadAt(b, 0)
	if err != nil {
		return err
	}

	f.mime = mimetype.Detect(b)

	return nil
}

func (f *ProcessableFile) FreeResources() error {
	err := f.File.Close()
	if err != nil {
		return err
	}

	return f.Remove()
}

func (f *ProcessableFile) Remove() error {
	return capyfs.FilesystemUtils.Remove(f.File.Name())
}

func (f *ProcessableFile) SetFileProcessingError(fileProcessingError FileProcessingError) {
	f.FileProcessingError = fileProcessingError
}

func (f *ProcessableFile) HasFileProcessingError() bool {
	return f.FileProcessingError != nil
}

func (f *ProcessableFile) GeneratedFilename() string {
	_ = f.loadMime()

	if f.mime == nil {
		return f.NanoID
	}

	return f.NanoID + f.mime.Extension()
}

func (f *ProcessableFile) OriginalFilename() string {
	return f.Metadata.OriginalFilename
}

func (f *ProcessableFile) AddOperationMetadata(key string, val interface{}) {
	if f.OperationMetadata == nil {
		f.OperationMetadata = make(map[string]interface{})
	}

	f.OperationMetadata[key] = val
}
