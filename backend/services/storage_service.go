package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

// UploadFileToSupabase uploads a file to Supabase Storage and returns the public URL
func UploadFileToSupabase(file multipart.File, filename string, bucketName string) (string, error) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_ANON_KEY")
	
	if supabaseURL == "" || supabaseKey == "" {
		return "", fmt.Errorf("SUPABASE_URL and SUPABASE_ANON_KEY must be set")
	}

	// Read file content
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Generate unique filename with timestamp
	name := fmt.Sprintf("%d_%s", time.Now().Unix(), filename)
	
	// Upload file to Supabase Storage
	uploadURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", supabaseURL, bucketName, name)
	
	req, err := http.NewRequest("POST", uploadURL, bytes.NewReader(fileBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+supabaseKey)
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("x-upsert", "true") // Overwrite if exists

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Return public URL
	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", supabaseURL, bucketName, name)
	log.Printf("File uploaded successfully: %s", publicURL)
	return publicURL, nil
}

// GetPublicURL returns the public URL for a file in Supabase Storage
func GetPublicURL(bucketName string, filename string) string {
	supabaseURL := os.Getenv("SUPABASE_URL")
	return fmt.Sprintf("%s/storage/v1/object/public/%s/%s", supabaseURL, bucketName, filename)
}

// ValidateFile validates file size and type
func ValidateFile(file multipart.File, maxSize int64, allowedTypes []string) error {
	// Check file size
	fileInfo, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		return fmt.Errorf("failed to get file size: %w", err)
	}
	file.Seek(0, io.SeekStart) // Reset to beginning

	if fileInfo > maxSize {
		return fmt.Errorf("file size exceeds maximum allowed size of %d bytes", maxSize)
	}

	// Note: File type validation would require reading file header
	// For now, we'll rely on frontend validation and file extension
	return nil
}

// CreateBucketIfNotExists creates a storage bucket if it doesn't exist
// Returns nil if service role key is not set (buckets should be created manually)
func CreateBucketIfNotExists(bucketName string, isPublic bool) error {
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseServiceKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	
	if supabaseURL == "" {
		return fmt.Errorf("SUPABASE_URL must be set")
	}
	
	// If service role key is not set, skip automatic bucket creation
	// User should create buckets manually via Supabase dashboard
	if supabaseServiceKey == "" {
		return nil
	}

	// Check if bucket exists
	checkURL := fmt.Sprintf("%s/storage/v1/bucket/%s", supabaseURL, bucketName)
	req, _ := http.NewRequest("HEAD", checkURL, nil)
	req.Header.Set("Authorization", "Bearer "+supabaseServiceKey)
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err == nil && resp.StatusCode == http.StatusOK {
		// Bucket exists
		return nil
	}

	// Create bucket
	createURL := fmt.Sprintf("%s/storage/v1/bucket", supabaseURL)
	payload := map[string]interface{}{
		"name":   bucketName,
		"public": isPublic,
	}
	
	jsonData, _ := json.Marshal(payload)
	req, _ = http.NewRequest("POST", createURL, bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+supabaseServiceKey)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err = client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create bucket with status %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("Bucket '%s' created successfully", bucketName)
	return nil
}

