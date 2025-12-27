package services

import (
	"io"
	"mc-admin/internal/clients/files"
)

// FileInfo is an alias for files.FileInfo for backward compatibility
type FileInfo = files.FileInfo

// FileService wraps MinecraftFilesClient to provide file operations
type FileSystemAccessor interface {
	GetAbsolutePath(path string) (string, error)
	ListFiles(path string) ([]FileInfo, error)
	ReadFile(path string) (string, error)
	CreateDirectory(path string) error
	Delete(path string) error
	SaveFile(path string, content string) error
	SaveFileStream(path string, r io.Reader) error
}

type FileService struct {
	client *FileSystemAccessor
}

// NewFileService creates a new FileService with the given base path and max display size
func NewFileService(client FileSystemAccessor) *FileService {
	return &FileService{
		client: &client,
	}
}

// GetAbsolutePath returns the absolute path for a given relative path
func (s *FileService) GetAbsolutePath(path string) (string, error) {
	return (*s.client).GetAbsolutePath(path)
}

// ListFiles returns a list of files and directories at the given path
func (s *FileService) ListFiles(path string) ([]FileInfo, error) {
	return (*s.client).ListFiles(path)
}

// ReadFile reads and returns the content of a file
func (s *FileService) ReadFile(path string) (string, error) {
	return (*s.client).ReadFile(path)
}

// CreateDirectory creates a directory at the given path
func (s *FileService) CreateDirectory(path string) error {
	return (*s.client).CreateDirectory(path)
}

// Delete removes a file or directory at the given path
func (s *FileService) Delete(path string) error {
	return (*s.client).Delete(path)
}

// SaveFile writes content to a file at the given path
func (s *FileService) SaveFile(path string, content string) error {
	return (*s.client).SaveFile(path, content)
}

// SaveFileStream writes content from a reader to a file at the given path
func (s *FileService) SaveFileStream(path string, r io.Reader) error {
	return (*s.client).SaveFileStream(path, r)
}
