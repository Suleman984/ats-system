package services

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

var storageClient = &http.Client{Timeout: 30 * time.Second}

// UploadFile stores data under a random object name and returns its public URL.
func UploadFile(data []byte, ext, contentType, bucket string) (string, error) {
	baseURL := strings.TrimRight(os.Getenv("SUPABASE_URL"), "/")
	serviceKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	if baseURL == "" || serviceKey == "" {
		return "", errors.New("SUPABASE_URL and SUPABASE_SERVICE_ROLE_KEY must be set")
	}

	object := uuid.NewString() + ext
	uploadURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", baseURL, bucket, object)

	req, err := http.NewRequest(http.MethodPost, uploadURL, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+serviceKey)
	req.Header.Set("apikey", serviceKey)
	req.Header.Set("Content-Type", contentType)

	resp, err := storageClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("upload request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return "", fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, body)
	}
	return fmt.Sprintf("%s/storage/v1/object/public/%s/%s", baseURL, bucket, object), nil
}
