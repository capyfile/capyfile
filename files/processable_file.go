package files

import (
	"capyfile/capyfs"
	"github.com/gabriel-vasile/mimetype"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"path/filepath"
)

const (
	cleanupPolicyNone   = 0
	cleanupPolicyKeep   = 1
	cleanupPolicyRemove = 2
)

// ProcessableFile The file that can be processed by the operations.
type ProcessableFile struct {
	NanoID string

	FileProcessingError FileProcessingError

	Metadata *ProcessableFileMetadata
	// OperationMetadata The metadata related to a specific operation.
	OperationMetadata map[string]interface{}

	// PreserveOriginalProcessableFile Whether the original file should be stored along with the processed one.
	PreserveOriginalProcessableFile bool
	// OriginalProcessableFile Sometimes we might need to preserve the original unmodified file
	// to store it along with the modified (processed file, like resized image) file.
	OriginalProcessableFile *ProcessableFile

	name string
	mime *mimetype.MIME
	// The thing with this parameter is that it should be set only once.
	// We should be careful with it because we don't want to remove the files
	// that has no clear indication of whether they should be removed or not.
	cleanupPolicy int
}

func NewProcessableFile(name string) ProcessableFile {
	return ProcessableFile{
		name:   name,
		NanoID: gonanoid.Must(),
		Metadata: &ProcessableFileMetadata{
			OriginalFilename: name,
		},
		PreserveOriginalProcessableFile: true,
	}
}

// ReplaceFile Replaces the file associated with the processable file.
// Here it also updates everything that is related to it, the things like MIME type.
func (f *ProcessableFile) ReplaceFile(name string) {
	if f.PreserveOriginalProcessableFile {
		if f.OriginalProcessableFile != nil {
			_ = f.FreeResources()
		} else {
			// If we want to preserve original file and there are no original file associated
			// with this instance, we can consider this instance as the original file.
			f.OriginalProcessableFile = &ProcessableFile{
				name: f.name,
				mime: f.mime,
			}
		}
	} else {
		_ = f.FreeResources()
	}

	f.name = name
	f.mime = nil
}

func (f *ProcessableFile) KeepOnFreeResources() {
	if f.cleanupPolicy != cleanupPolicyNone {
		return
	}

	f.cleanupPolicy = cleanupPolicyKeep
}

func (f *ProcessableFile) RemoveOnFreeResources() {
	if f.cleanupPolicy != cleanupPolicyNone {
		return
	}

	f.cleanupPolicy = cleanupPolicyRemove
}

func (f *ProcessableFile) Mime() (*mimetype.MIME, error) {
	err := f.loadMime()
	if err != nil {
		return nil, err
	}

	return f.mime, nil
}

func (f *ProcessableFile) loadMime() (err error) {
	if f.mime != nil {
		return nil
	}

	file, fileOpenErr := capyfs.Filesystem.Open(f.name)
	if fileOpenErr != nil {
		return err
	}

	f.mime, err = mimetype.DetectReader(file)

	return err
}

func (f *ProcessableFile) FreeResources() error {
	if f.cleanupPolicy == cleanupPolicyRemove {
		return f.Remove()
	}

	return nil
}

func (f *ProcessableFile) Remove() error {
	return capyfs.FilesystemUtils.Remove(f.name)
}

func (f *ProcessableFile) SetFileProcessingError(fileProcessingError FileProcessingError) {
	f.FileProcessingError = fileProcessingError
}

func (f *ProcessableFile) HasFileProcessingError() bool {
	return f.FileProcessingError != nil
}

// GeneratedFilename The basename is the NanoID of the file and the extension is the MIME type extension.
func (f *ProcessableFile) GeneratedFilename() string {
	_ = f.loadMime()

	if f.mime == nil {
		return f.NanoID
	}

	return f.NanoID + f.mime.Extension()
}

func (f *ProcessableFile) Name() string {
	return f.name
}

func (f *ProcessableFile) Filename() string {
	return filepath.Base(f.name)
}

func (f *ProcessableFile) FileAbsolutePath() (string, error) {
	return filepath.Abs(f.name)
}

// FileBasename The basename of the file (filename without extension).
func (f *ProcessableFile) FileBasename() string {
	return f.Filename()[0 : len(f.Filename())-len(f.FileExtension())]
}

// FileExtension The extension of the generated file.
func (f *ProcessableFile) FileExtension() string {
	return filepath.Ext(f.name)
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
