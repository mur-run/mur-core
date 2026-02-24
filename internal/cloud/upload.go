package cloud

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	DefaultWorkflowAPIURL = "https://mur-workflow-api.mur-run.workers.dev"
)

// UploadResponse is the response from the workflow upload API.
type UploadResponse struct {
	URL       string `json:"url"`
	Key       string `json:"key"`
	ExpiresAt string `json:"expires_at"`
}

// UploadResult contains the URL and key from a successful upload.
type UploadResult struct {
	URL string
	Key string
}

// UploadSessionData compresses and uploads session data to the workflow API,
// returning the shareable workflow URL.
func UploadSessionData(apiURL string, data []byte) (string, error) {
	if apiURL == "" {
		apiURL = DefaultWorkflowAPIURL
	}

	// Gzip the data
	var buf bytes.Buffer
	gz, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return "", fmt.Errorf("create gzip writer: %w", err)
	}
	if _, err := gz.Write(data); err != nil {
		return "", fmt.Errorf("gzip write: %w", err)
	}
	if err := gz.Close(); err != nil {
		return "", fmt.Errorf("gzip close: %w", err)
	}

	// POST to the upload endpoint
	req, err := http.NewRequest("POST", apiURL+"/upload", &buf)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("upload request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != 201 {
		var errResp struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
			return "", fmt.Errorf("upload failed: %s", errResp.Error)
		}
		return "", fmt.Errorf("upload failed with status %d", resp.StatusCode)
	}

	var result UploadResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	return result.URL, nil
}

// UploadSessionDataFull compresses and uploads session data to the workflow API,
// returning both the shareable URL and the session key.
func UploadSessionDataFull(apiURL string, data []byte) (*UploadResult, error) {
	if apiURL == "" {
		apiURL = DefaultWorkflowAPIURL
	}

	var buf bytes.Buffer
	gz, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return nil, fmt.Errorf("create gzip writer: %w", err)
	}
	if _, err := gz.Write(data); err != nil {
		return nil, fmt.Errorf("gzip write: %w", err)
	}
	if err := gz.Close(); err != nil {
		return nil, fmt.Errorf("gzip close: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL+"/upload", &buf)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("upload request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != 201 {
		var errResp struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
			return nil, fmt.Errorf("upload failed: %s", errResp.Error)
		}
		return nil, fmt.Errorf("upload failed with status %d", resp.StatusCode)
	}

	var result UploadResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return &UploadResult{URL: result.URL, Key: result.Key}, nil
}
