package api

import (
	"html"
	"mc-admin/internal/services"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func handleGetFiles(fileService *services.FileService) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Query("path")
		if path == "" {
			path = "/"
		}

		files, err := fileService.ListFiles(path)
		if err != nil {
			// If error (e.g. path not found), maybe redirect to root or show error
			// For now, let's just show the error in the template or JSON
			if c.GetHeader("X-Requested-With") == "XMLHttpRequest" {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			data := getCommonPageData(c)
			data["Error"] = err.Error()

			if c.GetHeader("HX-Request") == "true" {
				c.HTML(http.StatusOK, "files.html", data)
				return
			}

			data["ActiveModule"] = "files"
			c.HTML(http.StatusOK, "index.html", data)
			return
		}

		// If it's an AJAX request, return JSON
		if c.GetHeader("X-Requested-With") == "XMLHttpRequest" {
			c.JSON(http.StatusOK, gin.H{"files": files, "path": path})
			return
		}

		data := getCommonPageData(c)
		data["Files"] = files
		data["CurrentPath"] = path

		if c.GetHeader("HX-Request") == "true" {
			c.HTML(http.StatusOK, "files.html", data)
			return
		}

		// Otherwise render the template
		data["ActiveModule"] = "files"
		c.HTML(http.StatusOK, "index.html", data)
	}
}

func handleGetFileContent(fileService *services.FileService) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Query("path")
		content, err := fileService.ReadFile(path)
		if err != nil {
			c.String(http.StatusBadRequest, "Error: "+err.Error())
			return
		}
		// Escape HTML to prevent XSS and ensure correct display of code
		safeContent := html.EscapeString(content)
		c.String(http.StatusOK, safeContent)
	}
}

func handleCreateFile(fileService *services.FileService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Path    string `json:"path"`
			Content string `json:"content"`
			IsDir   bool   `json:"is_dir"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.IsDir {
			if err := fileService.CreateDirectory(req.Path); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		} else {
			if err := fileService.SaveFile(req.Path, req.Content); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	}
}

func handleDeleteFile(fileService *services.FileService) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Query("path")
		if err := fileService.Delete(path); err != nil {
			c.Header("HX-Trigger", `{"showToast": {"message": "Failed to delete file: `+err.Error()+`", "type": "error"}}`)
			c.String(http.StatusInternalServerError, "Error: "+err.Error())
			return
		}

		// Return updated file list
		parentDir := filepath.Dir(path)
		files, err := fileService.ListFiles(parentDir)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error listing files: "+err.Error())
			return
		}

		filename := filepath.Base(path)
		c.Header("HX-Trigger", `{"showToast": {"message": "File '`+filename+`' deleted successfully", "type": "success"}}`)
		data := getCommonPageData(c)
		data["Files"] = files
		data["CurrentPath"] = parentDir
		c.HTML(http.StatusOK, "files.html", data)
	}
}

func handleUploadFile(fileService *services.FileService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get path from form
		path := c.PostForm("path")
		if path == "" {
			path = "/"
		}

		// Get file
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
			return
		}

		// Open file
		src, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer src.Close()

		targetPath := path
		if targetPath != "/" && targetPath[len(targetPath)-1] != '/' {
			targetPath += "/"
		}
		targetPath += file.Filename

		if err := fileService.SaveFileStream(targetPath, src); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "success"})
	}
}

func handleDownloadFile(fileService *services.FileService) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Query("path")
		absPath, err := fileService.GetAbsolutePath(path)
		if err != nil {
			c.String(http.StatusBadRequest, "Error: "+err.Error())
			return
		}

		filename := filepath.Base(absPath)
		c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
		c.File(absPath)
	}
}
