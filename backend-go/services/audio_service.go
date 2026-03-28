package services

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type AudioService struct {
	uploadDir string
	maxBytes  int64
}

func NewAudioService(uploadDir string, maxBytes int64) *AudioService {
	_ = os.MkdirAll(uploadDir, 0755)
	return &AudioService{
		uploadDir: uploadDir,
		maxBytes:  maxBytes,
	}
}

func (s *AudioService) StoreAndRead(file *multipart.FileHeader) (string, string, []byte, error) {
	src, err := file.Open()
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	var buf bytes.Buffer
	if _, err := io.CopyN(&buf, src, 512); err != nil && err != io.EOF {
		return "", "", nil, fmt.Errorf("failed to read uploaded file header: %w", err)
	}
	headerBytes := buf.Bytes()
	mimeType := file.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = http.DetectContentType(headerBytes)
	}

	remaining, err := io.ReadAll(io.LimitReader(src, s.maxBytes-int64(len(headerBytes))))
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to read uploaded file: %w", err)
	}
	if int64(len(headerBytes)+len(remaining)) > s.maxBytes {
		return "", "", nil, fmt.Errorf("file too large")
	}
	if _, err := buf.Write(remaining); err != nil {
		return "", "", nil, fmt.Errorf("failed to buffer file: %w", err)
	}

	ext := filepath.Ext(file.Filename)
	if ext == "" {
		ext = ".bin"
	}
	filename := uuid.New().String() + ext
	destPath := filepath.Join(s.uploadDir, filename)
	if err := os.WriteFile(destPath, buf.Bytes(), 0644); err != nil {
		return "", "", nil, fmt.Errorf("failed to store uploaded file: %w", err)
	}

	return destPath, mimeType, buf.Bytes(), nil
}

