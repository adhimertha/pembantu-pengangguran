package services

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/ledongthuc/pdf"
)

type CVService struct {
	uploadDir string
}

func NewCVService(uploadDir string) *CVService {
	// Ensure upload directory exists
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.MkdirAll(uploadDir, 0755)
	}
	return &CVService{
		uploadDir: uploadDir,
	}
}

// ExtractText extracts text from a PDF file
func (s *CVService) ExtractText(file *multipart.FileHeader) (string, error) {
	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Create a temporary file to save the uploaded PDF
	tempFile, err := os.CreateTemp(s.uploadDir, "cv-*.pdf")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Copy the uploaded file to the temporary file
	if _, err := io.Copy(tempFile, src); err != nil {
		return "", fmt.Errorf("failed to copy file to temp: %w", err)
	}

	// Read text from PDF
	text, err := s.readPdf(tempFile.Name())
	if err != nil {
		return "", fmt.Errorf("failed to parse pdf: %w", err)
	}

	return text, nil
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

// SaveFile saves the uploaded file and returns the path
func (s *CVService) SaveFile(file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Use original filename with some randomness to avoid collisions
	filename := fmt.Sprintf("%d-%s", os.Getpid(), filepath.Base(file.Filename))
	destPath := filepath.Join(s.uploadDir, filename)

	dst, err := os.Create(destPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}

	return destPath, nil
}
