package controllers

import (
	"net/http"

	"backend-go/services"

	"github.com/gin-gonic/gin"
)

type CVController struct {
	cvService *services.CVService
}

func NewCVController(cvService *services.CVService) *CVController {
	return &CVController{
		cvService: cvService,
	}
}

// Upload handles POST /api/cv/upload
func (ctrl *CVController) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// Validate file type (e.g., PDF)
	if file.Header.Get("Content-Type") != "application/pdf" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only PDF files are supported currently"})
		return
	}

	// Extract text from PDF
	text, err := ctrl.cvService.ExtractText(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to extract text from CV"})
		return
	}

	// For simplicity, return the extracted text and file details
	// In a real app, we might store the text and file in a DB
	c.JSON(http.StatusOK, gin.H{
		"filename":       file.Filename,
		"extracted_text": text,
		"message":        "CV uploaded and parsed successfully",
	})
}
