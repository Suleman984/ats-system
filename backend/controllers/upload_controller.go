package controllers

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"ats-backend/services"

	"github.com/gin-gonic/gin"
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
	handleUpload(c, "resumes", cvTypes, "PDF, DOC or DOCX")
}

func UploadPortfolio(c *gin.Context) {
	handleUpload(c, "portfolios", portfolioTypes, "PDF, ZIP, RAR or 7Z")
}

// handleUpload checks extension and file signature before storing the file.
func handleUpload(c *gin.Context, bucket string, allowed map[string]fileKind, allowedLabel string) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize+1<<20)
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided or file exceeds 10 MB"})
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	kind, ok := allowed[ext]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Allowed: " + allowedLabel})
		return
	}

	data, err := io.ReadAll(io.LimitReader(file, maxUploadSize+1))
	if err != nil || len(data) > maxUploadSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File exceeds the 10 MB limit"})
		return
	}
	if !bytes.HasPrefix(data, kind.magic) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File content does not match its extension"})
		return
	}

	fileURL, err := services.UploadFile(data, ext, kind.contentType, bucket)
	if err != nil {
		log.Printf("upload to %s: %v", bucket, err)
		serverError(c, "Failed to upload file")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully", "file_url": fileURL})
}
