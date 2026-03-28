package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/ledongthuc/pdf"
)

type CVService struct {
	uploadDir string
	storage   *SupabaseStorage
	bucket    string
}

func NewCVService(uploadDir string, storage *SupabaseStorage, bucket string) *CVService {
	// Ensure upload directory exists
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.MkdirAll(uploadDir, 0755)
	}
	return &CVService{
		uploadDir: uploadDir,
		storage:   storage,
		bucket:    bucket,
	}
}

func (s *CVService) UploadAndExtractText(ctx context.Context, file *multipart.FileHeader) (string, string, error) {
	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return "", "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Create a temporary file to save the uploaded PDF
	tempFile, err := os.CreateTemp(s.uploadDir, "cv-*.pdf")
	if err != nil {
		return "", "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Copy the uploaded file to the temporary file
	if _, err := io.Copy(tempFile, src); err != nil {
		return "", "", fmt.Errorf("failed to copy file to temp: %w", err)
	}

	// Read text from PDF
	text, err := s.readPdf(tempFile.Name())
	if err != nil {
		return "", "", fmt.Errorf("failed to parse pdf: %w", err)
	}

	if s.storage == nil || s.bucket == "" {
		return text, "", nil
	}

	data, err := os.ReadFile(tempFile.Name())
	if err != nil {
		return "", "", fmt.Errorf("failed to read temp file: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext == "" {
		ext = ".pdf"
	}
	objectPath := "cv/" + uuid.New().String() + ext
	publicURL, _, err := s.storage.Upload(ctx, s.bucket, objectPath, "application/pdf", data)
	if err != nil {
		return "", "", err
	}

	return text, publicURL, nil
}

func (s *CVService) readPdf(path string) (string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		return "", err
	}

	_, err = buf.ReadFrom(b)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (s *CVService) SaveFile(_ *multipart.FileHeader) (string, error) {
	return "", fmt.Errorf("deprecated: use UploadAndExtractText")
}
