package files

import (
	"fmt"
	"io"
	"mc-admin/internal/config"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// FileInfo represents metadata about a file or directory
type FileInfo struct {
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	IsDir   bool   `json:"is_dir"`
	ModTime string `json:"mod_time"`
}

// MinecraftFilesClient handles filesystem operations for Minecraft server files
type MinecraftFilesClient struct {
	BasePath       string
	MaxDisplaySize int64
}

// NewMinecraftFilesClient creates a new MinecraftFilesClient with the given base path and max display size
func NewMinecraftFilesClient(basePath string, maxDisplaySize int64) MinecraftFilesClient {
	if basePath == "" {
		basePath = "/data"
	}
	if maxDisplaySize <= 0 {
		maxDisplaySize = 1024 * 1024 // Default 1MB
	}
	return MinecraftFilesClient{
		BasePath:       basePath,
		MaxDisplaySize: maxDisplaySize,
	}
}

// BuildMinecraftFilesClientFromEnv creates a MinecraftFilesClient from environment variables
func BuildMinecraftFilesClientFromEnv() MinecraftFilesClient {
	basePath := config.GetEnv("MC_DATA_PATH")
	if basePath == nil {
		basePath = new(string)
		*basePath = "/data"
	}

	maxDisplaySize := int64(1024 * 1024) // Default 1MB
	if maxSizeEnv := config.GetEnv("MC_MAX_DISPLAY_SIZE"); maxSizeEnv != nil {
		// Parse max size if provided (in bytes)
		var size int64
		if _, err := fmt.Sscanf(*maxSizeEnv, "%d", &size); err == nil && size > 0 {
			maxDisplaySize = size
		}
	}

	return NewMinecraftFilesClient(*basePath, maxDisplaySize)
}

// resolvePath performs security check to ensure path is within BasePath
func (c *MinecraftFilesClient) resolvePath(path string) (string, error) {
	// Clean the path to remove .. and .
	cleanPath := filepath.Clean(path)
	// Join with base path
	fullPath := filepath.Join(c.BasePath, cleanPath)

	absBase, err := filepath.Abs(c.BasePath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve base path: %w", err)
	}
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}

	if !strings.HasPrefix(absPath, absBase) {
		return "", fmt.Errorf("access denied: path outside base directory")
	}

	return absPath, nil
}

// GetAbsolutePath returns the absolute path for a given relative path
func (c *MinecraftFilesClient) GetAbsolutePath(path string) (string, error) {
	return c.resolvePath(path)
}

// ListFiles returns a list of files and directories at the given path
func (c *MinecraftFilesClient) ListFiles(path string) ([]FileInfo, error) {
	fullPath, err := c.resolvePath(path)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var files []FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		files = append(files, FileInfo{
			Name:    entry.Name(),
			Size:    info.Size(),
			IsDir:   entry.IsDir(),
			ModTime: info.ModTime().Format("2006-01-02 15:04:05"),
		})
	}
	return files, nil
}

// ReadFile reads and returns the content of a file
func (c *MinecraftFilesClient) ReadFile(path string) (string, error) {
	fullPath, err := c.resolvePath(path)
	if err != nil {
		return "", err
	}

	// Check if it's a directory
	info, err := os.Stat(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to stat file: %w", err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("cannot read: path is a directory")
	}

	var content []byte
	truncated := false

	if info.Size() > c.MaxDisplaySize {
		f, err := os.Open(fullPath)
		if err != nil {
			return "", fmt.Errorf("failed to open file: %w", err)
		}
		defer f.Close()

		content = make([]byte, c.MaxDisplaySize)
		n, err := f.Read(content)
		if err != nil && err != io.EOF {
			return "", fmt.Errorf("failed to read file: %w", err)
		}
		content = content[:n]
		truncated = true
	} else {
		content, err = os.ReadFile(fullPath)
		if err != nil {
			return "", fmt.Errorf("failed to read file: %w", err)
		}
	}

	if c.isBinary(content) {
		return "", fmt.Errorf("cannot display: binary file")
	}

	result := string(content)
	if truncated {
		result += fmt.Sprintf("\n\n... File truncated (showing first %d bytes of %d bytes) ...", c.MaxDisplaySize, info.Size())
	}

	return result, nil
}

// isBinary checks if the content appears to be binary data
func (c *MinecraftFilesClient) isBinary(content []byte) bool {
	if len(content) == 0 {
		return false
	}

	// Check first 512 bytes for null byte
	limit := 512
	if len(content) < limit {
		limit = len(content)
	}
	for i := 0; i < limit; i++ {
		if content[i] == 0 {
			return true
		}
	}

	contentType := http.DetectContentType(content)
	return !strings.HasPrefix(contentType, "text/") &&
		!strings.Contains(contentType, "json") &&
		!strings.Contains(contentType, "xml") &&
		!strings.Contains(contentType, "html")
}

// CreateDirectory creates a directory at the given path
func (c *MinecraftFilesClient) CreateDirectory(path string) error {
	fullPath, err := c.resolvePath(path)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	return nil
}

// Delete removes a file or directory at the given path
func (c *MinecraftFilesClient) Delete(path string) error {
	fullPath, err := c.resolvePath(path)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(fullPath); err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}
	return nil
}

// SaveFile writes content to a file at the given path
func (c *MinecraftFilesClient) SaveFile(path string, content string) error {
	fullPath, err := c.resolvePath(path)
	if err != nil {
		return err
	}
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}
	return nil
}

// SaveFileStream writes content from a reader to a file at the given path
func (c *MinecraftFilesClient) SaveFileStream(path string, r io.Reader) error {
	fullPath, err := c.resolvePath(path)
	if err != nil {
		return err
	}

	// Create file
	out, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	// Write content
	if _, err := io.Copy(out, r); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}

// Exists checks if a file or directory exists at the given path
func (c *MinecraftFilesClient) Exists(path string) (bool, error) {
	fullPath, err := c.resolvePath(path)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(fullPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("failed to check existence: %w", err)
}

// IsDirectory checks if the path is a directory
func (c *MinecraftFilesClient) IsDirectory(path string) (bool, error) {
	fullPath, err := c.resolvePath(path)
	if err != nil {
		return false, err
	}
	info, err := os.Stat(fullPath)
	if err != nil {
		return false, fmt.Errorf("failed to stat path: %w", err)
	}
	return info.IsDir(), nil
}

// GetFileInfo returns metadata about a file or directory
func (c *MinecraftFilesClient) GetFileInfo(path string) (*FileInfo, error) {
	fullPath, err := c.resolvePath(path)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path: %w", err)
	}
	return &FileInfo{
		Name:    info.Name(),
		Size:    info.Size(),
		IsDir:   info.IsDir(),
		ModTime: info.ModTime().Format("2006-01-02 15:04:05"),
	}, nil
}

// Rename renames or moves a file or directory
func (c *MinecraftFilesClient) Rename(oldPath, newPath string) error {
	oldFullPath, err := c.resolvePath(oldPath)
	if err != nil {
		return err
	}
	newFullPath, err := c.resolvePath(newPath)
	if err != nil {
		return err
	}
	if err := os.Rename(oldFullPath, newFullPath); err != nil {
		return fmt.Errorf("failed to rename: %w", err)
	}
	return nil
}

// Copy copies a file from source to destination
func (c *MinecraftFilesClient) Copy(srcPath, dstPath string) error {
	srcFullPath, err := c.resolvePath(srcPath)
	if err != nil {
		return err
	}
	dstFullPath, err := c.resolvePath(dstPath)
	if err != nil {
		return err
	}

	// Open source file
	src, err := os.Open(srcFullPath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(dstFullPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy content
	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	// Preserve permissions
	srcInfo, err := os.Stat(srcFullPath)
	if err == nil {
		if err := os.Chmod(dstFullPath, srcInfo.Mode()); err != nil {
			return fmt.Errorf("failed to set permissions: %w", err)
		}
	}

	return nil
}
