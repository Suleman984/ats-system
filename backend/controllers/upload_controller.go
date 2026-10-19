package controllers

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"unicode"

	"ats-backend/config"
	"ats-backend/models"
	"ats-backend/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const maxUploadSize = 10 << 20

type fileKind struct {
	magic       []byte
	contentType string
}

var (
	pdfKind  = fileKind{[]byte("%PDF"), "application/pdf"}
	zipMagic = []byte("PK\x03\x04")
	oleMagic = []byte{0xD0, 0xCF, 0x11, 0xE0}

	cvTypes = map[string]fileKind{
		".pdf":  pdfKind,
		".doc":  {oleMagic, "application/msword"},
		".docx": {zipMagic, "application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
	}
	portfolioTypes = map[string]fileKind{
		".pdf": pdfKind,
		".zip": {zipMagic, "application/zip"},
		".rar": {[]byte("Rar!"), "application/x-rar-compressed"},
		".7z":  {[]byte{'7', 'z', 0xBC, 0xAF, 0x27, 0x1C}, "application/x-7z-compressed"},
	}
)

func UploadCV(c *gin.Context) {
	handleUpload(c, models.UploadResume, services.BucketResumes, cvTypes, "PDF, DOC or DOCX")
}

func UploadPortfolio(c *gin.Context) {
	handleUpload(c, models.UploadPortfolio, services.BucketPortfolios, portfolioTypes, "PDF, ZIP, RAR or 7Z")
}

// handleUpload checks extension and file signature, stores the file privately under a random name
// and returns an upload ID (never a URL). An application must claim the ID within 24 hours.
func handleUpload(c *gin.Context, kind, bucket string, allowed map[string]fileKind, allowedLabel string) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize+1<<20)
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided or file exceeds 10 MB"})
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	fk, ok := allowed[ext]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Allowed: " + allowedLabel})
		return
	}

	data, err := io.ReadAll(io.LimitReader(file, maxUploadSize+1))
	if err != nil || len(data) > maxUploadSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File exceeds the 10 MB limit"})
		return
	}
	if len(data) == 0 || !bytes.HasPrefix(data, fk.magic) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File content does not match its extension"})
		return
	}

	if err := services.ScanForMalware(data); err != nil {
		if errors.Is(err, services.ErrInfected) {
			log.Printf("blocked infected %s upload: %v", kind, err)
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "This file was flagged as unsafe and can't be uploaded."})
			return
		}
		log.Printf("virus scan for %s upload: %v", kind, err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "File scanning is temporarily unavailable. Please try again shortly."})
		return
	}

	object := uuid.NewString() + ext
	if err := services.PutObject(bucket, object, data, fk.contentType); err != nil {
		log.Printf("upload to %s: %v", bucket, err)
		serverError(c, "Failed to upload file")
		return
	}
	upload := models.Upload{
		Kind: kind, ObjectPath: object, ContentType: fk.contentType,
		SizeBytes: len(data), OriginalName: safeFileName(header.Filename),
	}
	if err := config.DB.Create(&upload).Error; err != nil {
		go services.DeleteObjects(bucket, []string{object})
		serverError(c, "Failed to upload file")
		return
	}

	// Parse the CV now, while it's in memory, so submitting doesn't have to download it again.
	if kind == models.UploadResume {
		go func() {
			if text, err := services.ExtractTextFromBytes(data, fk.contentType, object); err == nil {
				config.DB.Model(&upload).Update("parsed_text", text)
			}
		}()
	}

	c.JSON(http.StatusCreated, gin.H{"upload_id": upload.ID, "file_name": upload.OriginalName, "size": upload.SizeBytes})
}

// safeFileName keeps a display-only name: no path, no control characters, bounded length.
func safeFileName(name string) string {
	name = filepath.Base(strings.ReplaceAll(name, "\\", "/"))
	name = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || r == '"' {
			return -1
		}
		return r
	}, name)
	if runes := []rune(name); len(runes) > 200 {
		name = string(runes[len(runes)-200:])
	}
	return name
}
