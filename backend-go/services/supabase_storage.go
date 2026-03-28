package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type SupabaseStorage struct {
	baseURL    string
	serviceKey string
	httpClient *http.Client
}

func NewSupabaseStorage(baseURL string, serviceKey string) *SupabaseStorage {
	return &SupabaseStorage{
		baseURL:    strings.TrimRight(baseURL, "/"),
		serviceKey: serviceKey,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

func (s *SupabaseStorage) Upload(ctx context.Context, bucket string, objectPath string, contentType string, data []byte) (string, string, error) {
	if s.baseURL == "" || s.serviceKey == "" {
		return "", "", fmt.Errorf("supabase storage is not configured")
	}
	if bucket == "" {
		return "", "", fmt.Errorf("bucket is required")
	}

	objectPath = strings.TrimLeft(objectPath, "/")

	u, err := url.Parse(s.baseURL + "/storage/v1/object/" + url.PathEscape(bucket) + "/" + objectPath)
	if err != nil {
		return "", "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(data))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+s.serviceKey)
	req.Header.Set("apikey", s.serviceKey)
	req.Header.Set("x-upsert", "true")
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	res, err := s.httpClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return "", "", fmt.Errorf("supabase storage upload failed: status=%d body=%s", res.StatusCode, string(body))
	}

	var out struct {
		Key string `json:"Key"`
	}
	_ = json.Unmarshal(body, &out)

	publicURL := s.baseURL + "/storage/v1/object/public/" + bucket + "/" + objectPath
	return publicURL, out.Key, nil
}

