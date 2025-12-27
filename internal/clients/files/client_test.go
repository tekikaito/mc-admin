package files

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupTestDir(t *testing.T) (string, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "mc-admin-files-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	cleanup := func() {
		os.RemoveAll(dir)
	}
	return dir, cleanup
}

func TestNewMinecraftFilesClient(t *testing.T) {
	tests := []struct {
		name               string
		basePath           string
		maxDisplaySize     int64
		expectedBasePath   string
		expectedMaxDisplay int64
	}{
		{
			name:               "with valid values",
			basePath:           "/custom/path",
			maxDisplaySize:     2048,
			expectedBasePath:   "/custom/path",
			expectedMaxDisplay: 2048,
		},
		{
			name:               "with empty base path",
			basePath:           "",
			maxDisplaySize:     1024,
			expectedBasePath:   "/data",
			expectedMaxDisplay: 1024,
		},
		{
			name:               "with zero max display size",
			basePath:           "/path",
			maxDisplaySize:     0,
			expectedBasePath:   "/path",
			expectedMaxDisplay: 1024 * 1024,
		},
		{
			name:               "with negative max display size",
			basePath:           "/path",
			maxDisplaySize:     -100,
			expectedBasePath:   "/path",
			expectedMaxDisplay: 1024 * 1024,
		},
		{
			name:               "with all defaults",
			basePath:           "",
			maxDisplaySize:     0,
			expectedBasePath:   "/data",
			expectedMaxDisplay: 1024 * 1024,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewMinecraftFilesClient(tt.basePath, tt.maxDisplaySize)
			if client.BasePath != tt.expectedBasePath {
				t.Errorf("BasePath = %q, want %q", client.BasePath, tt.expectedBasePath)
			}
			if client.MaxDisplaySize != tt.expectedMaxDisplay {
				t.Errorf("MaxDisplaySize = %d, want %d", client.MaxDisplaySize, tt.expectedMaxDisplay)
			}
		})
	}
}

func TestMinecraftFilesClient_resolvePath(t *testing.T) {
	baseDir, cleanup := setupTestDir(t)
	defer cleanup()

	client := NewMinecraftFilesClient(baseDir, 1024)

	tests := []struct {
		name      string
		path      string
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "valid path",
			path:    "subdir/file.txt",
			wantErr: false,
		},
		{
			name:    "root path",
			path:    "",
			wantErr: false,
		},
		{
			name:    "dot path",
			path:    ".",
			wantErr: false,
		},
		{
			name:      "path traversal attempt",
			path:      "../../../etc/passwd",
			wantErr:   true,
			errSubstr: "access denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved, err := client.GetAbsolutePath(tt.path)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("error = %q, want substring %q", err.Error(), tt.errSubstr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.HasPrefix(resolved, baseDir) {
				t.Errorf("resolved path %q does not start with base %q", resolved, baseDir)
			}
		})
	}
}

func TestMinecraftFilesClient_ListFiles(t *testing.T) {
	baseDir, cleanup := setupTestDir(t)
	defer cleanup()

	// Create test structure
	os.WriteFile(filepath.Join(baseDir, "file1.txt"), []byte("content1"), 0644)
	os.WriteFile(filepath.Join(baseDir, "file2.json"), []byte("{}"), 0644)
	os.MkdirAll(filepath.Join(baseDir, "subdir"), 0755)
	os.WriteFile(filepath.Join(baseDir, "subdir", "nested.txt"), []byte("nested"), 0644)

	client := NewMinecraftFilesClient(baseDir, 1024)

	t.Run("list root directory", func(t *testing.T) {
		files, err := client.ListFiles("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files) != 3 {
			t.Errorf("got %d files, want 3", len(files))
		}

		// Check we have expected entries
		names := make(map[string]bool)
		for _, f := range files {
			names[f.Name] = true
		}
		if !names["file1.txt"] || !names["file2.json"] || !names["subdir"] {
			t.Errorf("missing expected files, got: %v", names)
		}
	})

	t.Run("list subdirectory", func(t *testing.T) {
		files, err := client.ListFiles("subdir")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files) != 1 {
			t.Errorf("got %d files, want 1", len(files))
		}
		if files[0].Name != "nested.txt" {
			t.Errorf("got file %q, want nested.txt", files[0].Name)
		}
	})

	t.Run("list non-existent directory", func(t *testing.T) {
		_, err := client.ListFiles("nonexistent")
		if err == nil {
			t.Fatal("expected error for non-existent directory")
		}
	})

	t.Run("check file info fields", func(t *testing.T) {
		files, err := client.ListFiles("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, f := range files {
			if f.Name == "subdir" {
				if !f.IsDir {
					t.Error("subdir should be marked as directory")
				}
			}
			if f.Name == "file1.txt" {
				if f.IsDir {
					t.Error("file1.txt should not be marked as directory")
				}
				if f.Size != 8 {
					t.Errorf("file1.txt size = %d, want 8", f.Size)
				}
			}
			if f.ModTime == "" {
				t.Errorf("ModTime should not be empty for %s", f.Name)
			}
		}
	})
}

func TestMinecraftFilesClient_ReadFile(t *testing.T) {
	baseDir, cleanup := setupTestDir(t)
	defer cleanup()

	client := NewMinecraftFilesClient(baseDir, 100) // Small max size for truncation test

	// Create test files
	os.WriteFile(filepath.Join(baseDir, "small.txt"), []byte("hello world"), 0644)
	os.WriteFile(filepath.Join(baseDir, "large.txt"), []byte(strings.Repeat("x", 200)), 0644)
	os.WriteFile(filepath.Join(baseDir, "binary.bin"), []byte{0x00, 0x01, 0x02, 0x03}, 0644)
	os.MkdirAll(filepath.Join(baseDir, "dir"), 0755)

	t.Run("read small file", func(t *testing.T) {
		content, err := client.ReadFile("small.txt")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if content != "hello world" {
			t.Errorf("content = %q, want %q", content, "hello world")
		}
	})

	t.Run("read large file truncates", func(t *testing.T) {
		content, err := client.ReadFile("large.txt")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(content, "truncated") {
			t.Error("large file should contain truncation message")
		}
		if !strings.Contains(content, "100 bytes") {
			t.Error("truncation message should mention max size")
		}
	})

	t.Run("read binary file fails", func(t *testing.T) {
		_, err := client.ReadFile("binary.bin")
		if err == nil {
			t.Fatal("expected error for binary file")
		}
		if !strings.Contains(err.Error(), "binary") {
			t.Errorf("error = %q, should mention binary", err.Error())
		}
	})

	t.Run("read directory fails", func(t *testing.T) {
		_, err := client.ReadFile("dir")
		if err == nil {
			t.Fatal("expected error for directory")
		}
		if !strings.Contains(err.Error(), "directory") {
			t.Errorf("error = %q, should mention directory", err.Error())
		}
	})

	t.Run("read non-existent file", func(t *testing.T) {
		_, err := client.ReadFile("nonexistent.txt")
		if err == nil {
			t.Fatal("expected error for non-existent file")
		}
	})
}

func TestMinecraftFilesClient_SaveFile(t *testing.T) {
	baseDir, cleanup := setupTestDir(t)
	defer cleanup()

	client := NewMinecraftFilesClient(baseDir, 1024)

	t.Run("save new file", func(t *testing.T) {
		err := client.SaveFile("new.txt", "new content")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		content, err := os.ReadFile(filepath.Join(baseDir, "new.txt"))
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		if string(content) != "new content" {
			t.Errorf("content = %q, want %q", string(content), "new content")
		}
	})

	t.Run("overwrite existing file", func(t *testing.T) {
		os.WriteFile(filepath.Join(baseDir, "existing.txt"), []byte("old"), 0644)

		err := client.SaveFile("existing.txt", "new")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		content, err := os.ReadFile(filepath.Join(baseDir, "existing.txt"))
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		if string(content) != "new" {
			t.Errorf("content = %q, want %q", string(content), "new")
		}
	})

	t.Run("save with path traversal fails", func(t *testing.T) {
		err := client.SaveFile("../outside.txt", "content")
		if err == nil {
			t.Fatal("expected error for path traversal")
		}
	})
}

func TestMinecraftFilesClient_SaveFileStream(t *testing.T) {
	baseDir, cleanup := setupTestDir(t)
	defer cleanup()

	client := NewMinecraftFilesClient(baseDir, 1024)

	t.Run("save from reader", func(t *testing.T) {
		reader := strings.NewReader("streamed content")
		err := client.SaveFileStream("streamed.txt", reader)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		content, err := os.ReadFile(filepath.Join(baseDir, "streamed.txt"))
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		if string(content) != "streamed content" {
			t.Errorf("content = %q, want %q", string(content), "streamed content")
		}
	})
}

func TestMinecraftFilesClient_CreateDirectory(t *testing.T) {
	baseDir, cleanup := setupTestDir(t)
	defer cleanup()

	client := NewMinecraftFilesClient(baseDir, 1024)

	t.Run("create single directory", func(t *testing.T) {
		err := client.CreateDirectory("newdir")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		info, err := os.Stat(filepath.Join(baseDir, "newdir"))
		if err != nil {
			t.Fatalf("directory not created: %v", err)
		}
		if !info.IsDir() {
			t.Error("created path is not a directory")
		}
	})

	t.Run("create nested directories", func(t *testing.T) {
		err := client.CreateDirectory("a/b/c")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		info, err := os.Stat(filepath.Join(baseDir, "a/b/c"))
		if err != nil {
			t.Fatalf("nested directory not created: %v", err)
		}
		if !info.IsDir() {
			t.Error("created path is not a directory")
		}
	})

	t.Run("create existing directory succeeds", func(t *testing.T) {
		os.MkdirAll(filepath.Join(baseDir, "existing"), 0755)
		err := client.CreateDirectory("existing")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestMinecraftFilesClient_Delete(t *testing.T) {
	baseDir, cleanup := setupTestDir(t)
	defer cleanup()

	client := NewMinecraftFilesClient(baseDir, 1024)

	t.Run("delete file", func(t *testing.T) {
		os.WriteFile(filepath.Join(baseDir, "todelete.txt"), []byte("delete me"), 0644)

		err := client.Delete("todelete.txt")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, err := os.Stat(filepath.Join(baseDir, "todelete.txt")); !os.IsNotExist(err) {
			t.Error("file should be deleted")
		}
	})

	t.Run("delete directory recursively", func(t *testing.T) {
		os.MkdirAll(filepath.Join(baseDir, "deldir/sub"), 0755)
		os.WriteFile(filepath.Join(baseDir, "deldir/sub/file.txt"), []byte("content"), 0644)

		err := client.Delete("deldir")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, err := os.Stat(filepath.Join(baseDir, "deldir")); !os.IsNotExist(err) {
			t.Error("directory should be deleted")
		}
	})

	t.Run("delete non-existent succeeds", func(t *testing.T) {
		err := client.Delete("nonexistent")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestMinecraftFilesClient_Exists(t *testing.T) {
	baseDir, cleanup := setupTestDir(t)
	defer cleanup()

	client := NewMinecraftFilesClient(baseDir, 1024)

	os.WriteFile(filepath.Join(baseDir, "exists.txt"), []byte("content"), 0644)
	os.MkdirAll(filepath.Join(baseDir, "existsdir"), 0755)

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"existing file", "exists.txt", true},
		{"existing directory", "existsdir", true},
		{"non-existent file", "missing.txt", false},
		{"non-existent directory", "missingdir", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := client.Exists(tt.path)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if exists != tt.expected {
				t.Errorf("Exists(%q) = %v, want %v", tt.path, exists, tt.expected)
			}
		})
	}
}

func TestMinecraftFilesClient_IsDirectory(t *testing.T) {
	baseDir, cleanup := setupTestDir(t)
	defer cleanup()

	client := NewMinecraftFilesClient(baseDir, 1024)

	os.WriteFile(filepath.Join(baseDir, "file.txt"), []byte("content"), 0644)
	os.MkdirAll(filepath.Join(baseDir, "dir"), 0755)

	t.Run("file is not directory", func(t *testing.T) {
		isDir, err := client.IsDirectory("file.txt")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if isDir {
			t.Error("file should not be a directory")
		}
	})

	t.Run("directory is directory", func(t *testing.T) {
		isDir, err := client.IsDirectory("dir")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isDir {
			t.Error("dir should be a directory")
		}
	})

	t.Run("non-existent path errors", func(t *testing.T) {
		_, err := client.IsDirectory("nonexistent")
		if err == nil {
			t.Fatal("expected error for non-existent path")
		}
	})
}

func TestMinecraftFilesClient_GetFileInfo(t *testing.T) {
	baseDir, cleanup := setupTestDir(t)
	defer cleanup()

	client := NewMinecraftFilesClient(baseDir, 1024)

	os.WriteFile(filepath.Join(baseDir, "info.txt"), []byte("test content"), 0644)
	os.MkdirAll(filepath.Join(baseDir, "infodir"), 0755)

	t.Run("get file info", func(t *testing.T) {
		info, err := client.GetFileInfo("info.txt")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if info.Name != "info.txt" {
			t.Errorf("Name = %q, want %q", info.Name, "info.txt")
		}
		if info.Size != 12 {
			t.Errorf("Size = %d, want 12", info.Size)
		}
		if info.IsDir {
			t.Error("IsDir should be false")
		}
		if info.ModTime == "" {
			t.Error("ModTime should not be empty")
		}
	})

	t.Run("get directory info", func(t *testing.T) {
		info, err := client.GetFileInfo("infodir")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if info.Name != "infodir" {
			t.Errorf("Name = %q, want %q", info.Name, "infodir")
		}
		if !info.IsDir {
			t.Error("IsDir should be true")
		}
	})

	t.Run("non-existent path errors", func(t *testing.T) {
		_, err := client.GetFileInfo("nonexistent")
		if err == nil {
			t.Fatal("expected error for non-existent path")
		}
	})
}

func TestMinecraftFilesClient_Rename(t *testing.T) {
	baseDir, cleanup := setupTestDir(t)
	defer cleanup()

	client := NewMinecraftFilesClient(baseDir, 1024)

	t.Run("rename file", func(t *testing.T) {
		os.WriteFile(filepath.Join(baseDir, "old.txt"), []byte("content"), 0644)

		err := client.Rename("old.txt", "new.txt")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, err := os.Stat(filepath.Join(baseDir, "old.txt")); !os.IsNotExist(err) {
			t.Error("old file should not exist")
		}
		content, err := os.ReadFile(filepath.Join(baseDir, "new.txt"))
		if err != nil {
			t.Fatalf("new file not found: %v", err)
		}
		if string(content) != "content" {
			t.Errorf("content = %q, want %q", string(content), "content")
		}
	})

	t.Run("move file to subdirectory", func(t *testing.T) {
		os.WriteFile(filepath.Join(baseDir, "tomove.txt"), []byte("moveme"), 0644)
		os.MkdirAll(filepath.Join(baseDir, "subdir"), 0755)

		err := client.Rename("tomove.txt", "subdir/moved.txt")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		content, err := os.ReadFile(filepath.Join(baseDir, "subdir/moved.txt"))
		if err != nil {
			t.Fatalf("moved file not found: %v", err)
		}
		if string(content) != "moveme" {
			t.Errorf("content = %q, want %q", string(content), "moveme")
		}
	})

	t.Run("rename non-existent file fails", func(t *testing.T) {
		err := client.Rename("nonexistent.txt", "new.txt")
		if err == nil {
			t.Fatal("expected error for non-existent file")
		}
	})
}

func TestMinecraftFilesClient_Copy(t *testing.T) {
	baseDir, cleanup := setupTestDir(t)
	defer cleanup()

	client := NewMinecraftFilesClient(baseDir, 1024)

	t.Run("copy file", func(t *testing.T) {
		os.WriteFile(filepath.Join(baseDir, "source.txt"), []byte("source content"), 0644)

		err := client.Copy("source.txt", "dest.txt")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Check source still exists
		srcContent, err := os.ReadFile(filepath.Join(baseDir, "source.txt"))
		if err != nil {
			t.Fatalf("source file missing: %v", err)
		}
		if string(srcContent) != "source content" {
			t.Errorf("source content = %q, want %q", string(srcContent), "source content")
		}

		// Check destination was created
		dstContent, err := os.ReadFile(filepath.Join(baseDir, "dest.txt"))
		if err != nil {
			t.Fatalf("destination file missing: %v", err)
		}
		if string(dstContent) != "source content" {
			t.Errorf("destination content = %q, want %q", string(dstContent), "source content")
		}
	})

	t.Run("copy to subdirectory", func(t *testing.T) {
		os.WriteFile(filepath.Join(baseDir, "tocopy.txt"), []byte("copy me"), 0644)
		os.MkdirAll(filepath.Join(baseDir, "copydir"), 0755)

		err := client.Copy("tocopy.txt", "copydir/copied.txt")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		content, err := os.ReadFile(filepath.Join(baseDir, "copydir/copied.txt"))
		if err != nil {
			t.Fatalf("copied file not found: %v", err)
		}
		if string(content) != "copy me" {
			t.Errorf("content = %q, want %q", string(content), "copy me")
		}
	})

	t.Run("copy non-existent file fails", func(t *testing.T) {
		err := client.Copy("nonexistent.txt", "dest.txt")
		if err == nil {
			t.Fatal("expected error for non-existent file")
		}
	})
}

func TestMinecraftFilesClient_isBinary(t *testing.T) {
	client := NewMinecraftFilesClient("/tmp", 1024)

	tests := []struct {
		name     string
		content  []byte
		expected bool
	}{
		{"empty", []byte{}, false},
		{"text", []byte("hello world"), false},
		{"json", []byte(`{"key": "value"}`), false},
		{"xml", []byte(`<?xml version="1.0"?><root/>`), false},
		{"html", []byte(`<!DOCTYPE html><html></html>`), false},
		{"null byte", []byte{0x00, 0x01, 0x02}, true},
		{"null in middle", []byte("hello\x00world"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.isBinary(tt.content)
			if result != tt.expected {
				t.Errorf("isBinary(%v) = %v, want %v", tt.content, result, tt.expected)
			}
		})
	}
}

func TestMinecraftFilesClient_ReadFile_EmptyFile(t *testing.T) {
	baseDir, cleanup := setupTestDir(t)
	defer cleanup()

	client := NewMinecraftFilesClient(baseDir, 1024)

	os.WriteFile(filepath.Join(baseDir, "empty.txt"), []byte{}, 0644)

	content, err := client.ReadFile("empty.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != "" {
		t.Errorf("content = %q, want empty string", content)
	}
}
