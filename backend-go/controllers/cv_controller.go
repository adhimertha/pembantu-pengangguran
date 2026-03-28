package controllers

import (
	"net/http"
	"strings"

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

	if !strings.HasSuffix(strings.ToLower(file.Filename), ".pdf") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only PDF files are supported currently"})
		return
	}

	text, fileURL, err := ctrl.cvService.UploadAndExtractText(c.Request.Context(), file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to extract text from CV"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"filename":       file.Filename,
		"file_url":       fileURL,
		"extracted_text": text,
		"message":        "CV uploaded and parsed successfully",
	})
}
