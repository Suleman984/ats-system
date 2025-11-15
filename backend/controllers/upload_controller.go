package controllers

import (
	"ats-backend/services"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	maxFileSize      = 10 * 1024 * 1024 // 10MB
	cvBucketName      = "resumes"
	portfolioBucketName = "portfolios"
)

// UploadCV handles CV/resume file upload
func UploadCV(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}
	defer file.Close()

	// Validate file type
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedExts := []string{".pdf", ".doc", ".docx"}
	allowed := false
	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			allowed = true
			break
		}
	}
	if !allowed {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid file type. Only PDF, DOC, and DOCX files are allowed",
		})
		return
	}

	// Validate file size
	if header.Size > maxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("File size exceeds maximum allowed size of %d MB", maxFileSize/(1024*1024)),
		})
		return
	}

	// Upload to Supabase Storage
	publicURL, err := services.UploadFileToSupabase(file, header.Filename, cvBucketName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to upload file",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "File uploaded successfully",
		"file_url": publicURL,
	})
}

// UploadPortfolio handles portfolio file upload
func UploadPortfolio(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}
	defer file.Close()

	// Validate file type
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedExts := []string{".pdf", ".zip", ".rar", ".7z"}
	allowed := false
	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			allowed = true
			break
		}
	}
	if !allowed {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid file type. Only PDF, ZIP, RAR, and 7Z files are allowed",
		})
		return
	}

	// Validate file size
	if header.Size > maxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("File size exceeds maximum allowed size of %d MB", maxFileSize/(1024*1024)),
		})
		return
	}

	// Upload to Supabase Storage
	publicURL, err := services.UploadFileToSupabase(file, header.Filename, portfolioBucketName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to upload file",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "File uploaded successfully",
		"file_url": publicURL,
	})
}

// GetFileURL returns the public URL for a file (if URL is provided, returns as-is)
func GetFileURL(c *gin.Context) {
	var req struct {
		URL      string `json:"url"`
		Filename string `json:"filename"`
		Bucket   string `json:"bucket"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// If URL is provided and is a valid URL, return it as-is
	if req.URL != "" && (strings.HasPrefix(req.URL, "http://") || strings.HasPrefix(req.URL, "https://")) {
		c.JSON(http.StatusOK, gin.H{"file_url": req.URL})
		return
	}

	// Otherwise, construct URL from bucket and filename
	if req.Bucket == "" || req.Filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Either URL or bucket+filename must be provided"})
		return
	}

	publicURL := services.GetPublicURL(req.Bucket, req.Filename)
	c.JSON(http.StatusOK, gin.H{"file_url": publicURL})
}

