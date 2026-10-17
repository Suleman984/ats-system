package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// Private Supabase Storage buckets. Objects are only reachable with the service-role key, which
// never leaves the API, or through the short-lived signed URLs it issues.
const (
	BucketResumes    = "resumes"
	BucketPortfolios = "portfolios"
)

var storageClient = &http.Client{Timeout: 30 * time.Second}

func storageBase() (string, string, error) {
	baseURL := strings.TrimRight(os.Getenv("SUPABASE_URL"), "/")
	serviceKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	if baseURL == "" || serviceKey == "" {
		return "", "", errors.New("SUPABASE_URL and SUPABASE_SERVICE_ROLE_KEY must be set")
	}
	return baseURL + "/storage/v1", serviceKey, nil
}

func storageRequest(method, path string, body io.Reader, contentType string) (*http.Response, error) {
	base, key, err := storageBase()
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, base+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("apikey", key)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	resp, err := storageClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		resp.Body.Close()
		return nil, fmt.Errorf("storage %s %s returned %d: %s", method, path, resp.StatusCode, detail)
	}
	return resp, nil
}

func objectPath(bucket, object string) string {
	return "/object/" + bucket + "/" + url.PathEscape(object)
}

// PutObject stores a new object; an existing name is an error rather than an overwrite.
func PutObject(bucket, object string, data []byte, contentType string) error {
	resp, err := storageRequest(http.MethodPost, objectPath(bucket, object), bytes.NewReader(data), contentType)
	if err != nil {
		return err
	}
	return resp.Body.Close()
}

// GetObject reads a stored file, capped at the upload size limit.
func GetObject(bucket, object string) ([]byte, error) {
	resp, err := storageRequest(http.MethodGet, objectPath(bucket, object), nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(io.LimitReader(resp.Body, maxDownloadBytes+1))
}

// SignedURL returns a link to a private object that works for ttl. downloadName, if set, makes the
// browser save the file under that name instead of rendering it.
func SignedURL(bucket, object string, ttl time.Duration, downloadName string) (string, error) {
	body, _ := json.Marshal(map[string]int{"expiresIn": int(ttl.Seconds())})
	resp, err := storageRequest(http.MethodPost, "/object/sign/"+bucket+"/"+url.PathEscape(object), bytes.NewReader(body), "application/json")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var out struct {
		SignedURL string `json:"signedURL"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil || out.SignedURL == "" {
		return "", fmt.Errorf("storage sign: unexpected response: %v", err)
	}
	base, _, _ := storageBase()
	link := base + out.SignedURL
	if downloadName != "" {
		link += "&download=" + url.QueryEscape(downloadName)
	}
	return link, nil
}

// DeleteObjects removes objects from a bucket; missing ones are ignored by Storage.
func DeleteObjects(bucket string, objects []string) error {
	if len(objects) == 0 {
		return nil
	}
	body, _ := json.Marshal(map[string][]string{"prefixes": objects})
	resp, err := storageRequest(http.MethodDelete, "/object/"+bucket, bytes.NewReader(body), "application/json")
	if err != nil {
		return err
	}
	return resp.Body.Close()
}
