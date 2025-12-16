package services

import (
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
)

// fakeFileSystemAccessor is a mock implementation of FileSystemAccessor for testing
type fakeFileSystemAccessor struct {
	absPathResult   string
	absPathErr      error
	listFilesResult []FileInfo
	listFilesErr    error
	readFileResult  string
	readFileErr     error
	createDirErr    error
	deleteErr       error
	saveFileErr     error
	saveStreamErr   error

	// Track calls
	absPathCalls    []string
	listFilesCalls  []string
	readFileCalls   []string
	createDirCalls  []string
	deleteCalls     []string
	saveFileCalls   []struct{ path, content string }
	saveStreamCalls []string
}

func (f *fakeFileSystemAccessor) GetAbsolutePath(path string) (string, error) {
	f.absPathCalls = append(f.absPathCalls, path)
	return f.absPathResult, f.absPathErr
}

func (f *fakeFileSystemAccessor) ListFiles(path string) ([]FileInfo, error) {
	f.listFilesCalls = append(f.listFilesCalls, path)
	return f.listFilesResult, f.listFilesErr
}

func (f *fakeFileSystemAccessor) ReadFile(path string) (string, error) {
	f.readFileCalls = append(f.readFileCalls, path)
	return f.readFileResult, f.readFileErr
}

func (f *fakeFileSystemAccessor) CreateDirectory(path string) error {
	f.createDirCalls = append(f.createDirCalls, path)
	return f.createDirErr
}

func (f *fakeFileSystemAccessor) Delete(path string) error {
	f.deleteCalls = append(f.deleteCalls, path)
	return f.deleteErr
}

func (f *fakeFileSystemAccessor) SaveFile(path string, content string) error {
	f.saveFileCalls = append(f.saveFileCalls, struct{ path, content string }{path, content})
	return f.saveFileErr
}

func (f *fakeFileSystemAccessor) SaveFileStream(path string, r io.Reader) error {
	f.saveStreamCalls = append(f.saveStreamCalls, path)
	return f.saveStreamErr
}

func TestFileService_GetAbsolutePath(t *testing.T) {
	tests := []struct {
		name       string
		inputPath  string
		wantResult string
		wantErr    bool
		absPathErr error
	}{
		{
			name:       "valid path",
			inputPath:  "world/level.dat",
			wantResult: "/data/world/level.dat",
		},
		{
			name:       "root path",
			inputPath:  "",
			wantResult: "/data",
		},
		{
			name:       "path resolution error",
			inputPath:  "../etc/passwd",
			wantErr:    true,
			absPathErr: errors.New("access denied: path outside base directory"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeFileSystemAccessor{
				absPathResult: tt.wantResult,
				absPathErr:    tt.absPathErr,
			}
			svc := NewFileService(fake)

			got, err := svc.GetAbsolutePath(tt.inputPath)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.wantResult {
				t.Fatalf("GetAbsolutePath() = %q, want %q", got, tt.wantResult)
			}

			if len(fake.absPathCalls) != 1 || fake.absPathCalls[0] != tt.inputPath {
				t.Fatalf("GetAbsolutePath called with %v, want %q", fake.absPathCalls, tt.inputPath)
			}
		})
	}
}

func TestFileService_ListFiles(t *testing.T) {
	tests := []struct {
		name       string
		inputPath  string
		wantResult []FileInfo
		wantErr    bool
		listErr    error
	}{
		{
			name:      "list files in directory",
			inputPath: "world",
			wantResult: []FileInfo{
				{Name: "level.dat", Size: 1024, IsDir: false, ModTime: "2024-01-01 12:00:00"},
				{Name: "region", Size: 0, IsDir: true, ModTime: "2024-01-01 12:00:00"},
			},
		},
		{
			name:       "empty directory",
			inputPath:  "empty",
			wantResult: []FileInfo{},
		},
		{
			name:      "directory not found",
			inputPath: "nonexistent",
			wantErr:   true,
			listErr:   errors.New("failed to read directory: no such file or directory"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeFileSystemAccessor{
				listFilesResult: tt.wantResult,
				listFilesErr:    tt.listErr,
			}
			svc := NewFileService(fake)

			got, err := svc.ListFiles(tt.inputPath)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Fatalf("ListFiles() = %#v, want %#v", got, tt.wantResult)
			}

			if len(fake.listFilesCalls) != 1 || fake.listFilesCalls[0] != tt.inputPath {
				t.Fatalf("ListFiles called with %v, want %q", fake.listFilesCalls, tt.inputPath)
			}
		})
	}
}

func TestFileService_ReadFile(t *testing.T) {
	tests := []struct {
		name       string
		inputPath  string
		wantResult string
		wantErr    bool
		readErr    error
	}{
		{
			name:       "read text file",
			inputPath:  "server.properties",
			wantResult: "server-port=25565\nmotd=A Minecraft Server\n",
		},
		{
			name:       "read json file",
			inputPath:  "whitelist.json",
			wantResult: `[{"uuid":"123","name":"Steve"}]`,
		},
		{
			name:      "file not found",
			inputPath: "nonexistent.txt",
			wantErr:   true,
			readErr:   errors.New("failed to stat file: no such file or directory"),
		},
		{
			name:      "read directory error",
			inputPath: "world",
			wantErr:   true,
			readErr:   errors.New("cannot read: path is a directory"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeFileSystemAccessor{
				readFileResult: tt.wantResult,
				readFileErr:    tt.readErr,
			}
			svc := NewFileService(fake)

			got, err := svc.ReadFile(tt.inputPath)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.wantResult {
				t.Fatalf("ReadFile() = %q, want %q", got, tt.wantResult)
			}

			if len(fake.readFileCalls) != 1 || fake.readFileCalls[0] != tt.inputPath {
				t.Fatalf("ReadFile called with %v, want %q", fake.readFileCalls, tt.inputPath)
			}
		})
	}
}

func TestFileService_CreateDirectory(t *testing.T) {
	tests := []struct {
		name      string
		inputPath string
		wantErr   bool
		createErr error
	}{
		{
			name:      "create directory",
			inputPath: "world/new_folder",
		},
		{
			name:      "create nested directory",
			inputPath: "plugins/MyPlugin/config",
		},
		{
			name:      "directory already exists",
			inputPath: "world",
			wantErr:   true,
			createErr: errors.New("directory already exists"),
		},
		{
			name:      "permission denied",
			inputPath: "restricted/folder",
			wantErr:   true,
			createErr: errors.New("permission denied"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeFileSystemAccessor{
				createDirErr: tt.createErr,
			}
			svc := NewFileService(fake)

			err := svc.CreateDirectory(tt.inputPath)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(fake.createDirCalls) != 1 || fake.createDirCalls[0] != tt.inputPath {
				t.Fatalf("CreateDirectory called with %v, want %q", fake.createDirCalls, tt.inputPath)
			}
		})
	}
}

func TestFileService_Delete(t *testing.T) {
	tests := []struct {
		name      string
		inputPath string
		wantErr   bool
		deleteErr error
	}{
		{
			name:      "delete file",
			inputPath: "old_config.yml",
		},
		{
			name:      "delete directory",
			inputPath: "temp_folder",
		},
		{
			name:      "file not found",
			inputPath: "nonexistent.txt",
			wantErr:   true,
			deleteErr: errors.New("no such file or directory"),
		},
		{
			name:      "permission denied",
			inputPath: "protected_file.txt",
			wantErr:   true,
			deleteErr: errors.New("permission denied"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeFileSystemAccessor{
				deleteErr: tt.deleteErr,
			}
			svc := NewFileService(fake)

			err := svc.Delete(tt.inputPath)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(fake.deleteCalls) != 1 || fake.deleteCalls[0] != tt.inputPath {
				t.Fatalf("Delete called with %v, want %q", fake.deleteCalls, tt.inputPath)
			}
		})
	}
}

func TestFileService_SaveFile(t *testing.T) {
	tests := []struct {
		name        string
		inputPath   string
		content     string
		wantErr     bool
		saveFileErr error
	}{
		{
			name:      "save text file",
			inputPath: "server.properties",
			content:   "server-port=25565\nmotd=Updated Server\n",
		},
		{
			name:      "save json file",
			inputPath: "whitelist.json",
			content:   `[{"uuid":"123","name":"Steve"},{"uuid":"456","name":"Alex"}]`,
		},
		{
			name:      "save empty file",
			inputPath: "empty.txt",
			content:   "",
		},
		{
			name:        "permission denied",
			inputPath:   "readonly.txt",
			content:     "test content",
			wantErr:     true,
			saveFileErr: errors.New("permission denied"),
		},
		{
			name:        "disk full",
			inputPath:   "large_file.dat",
			content:     "content",
			wantErr:     true,
			saveFileErr: errors.New("no space left on device"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeFileSystemAccessor{
				saveFileErr: tt.saveFileErr,
			}
			svc := NewFileService(fake)

			err := svc.SaveFile(tt.inputPath, tt.content)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(fake.saveFileCalls) != 1 {
				t.Fatalf("SaveFile called %d times, want 1", len(fake.saveFileCalls))
			}
			call := fake.saveFileCalls[0]
			if call.path != tt.inputPath || call.content != tt.content {
				t.Fatalf("SaveFile called with (%q, %q), want (%q, %q)", call.path, call.content, tt.inputPath, tt.content)
			}
		})
	}
}

func TestFileService_SaveFileStream(t *testing.T) {
	tests := []struct {
		name          string
		inputPath     string
		content       string
		wantErr       bool
		saveStreamErr error
	}{
		{
			name:      "save file from stream",
			inputPath: "upload.txt",
			content:   "streamed content",
		},
		{
			name:      "save large file from stream",
			inputPath: "world/region/r.0.0.mca",
			content:   strings.Repeat("binary data ", 1000),
		},
		{
			name:          "permission denied",
			inputPath:     "readonly.txt",
			content:       "test content",
			wantErr:       true,
			saveStreamErr: errors.New("permission denied"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeFileSystemAccessor{
				saveStreamErr: tt.saveStreamErr,
			}
			svc := NewFileService(fake)

			reader := strings.NewReader(tt.content)
			err := svc.SaveFileStream(tt.inputPath, reader)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(fake.saveStreamCalls) != 1 || fake.saveStreamCalls[0] != tt.inputPath {
				t.Fatalf("SaveFileStream called with %v, want %q", fake.saveStreamCalls, tt.inputPath)
			}
		})
	}
}

func TestNewFileService(t *testing.T) {
	fake := &fakeFileSystemAccessor{}
	svc := NewFileService(fake)

	if svc == nil {
		t.Fatal("NewFileService returned nil")
	}
	if svc.client == nil {
		t.Fatal("FileService.client is nil")
	}
}
