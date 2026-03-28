package controllers

import (
	"net/http"

	"backend-go/services"

	"github.com/gin-gonic/gin"
)

type AudioController struct {
	aiService    services.LLM
	audioService *services.AudioService
}

func NewAudioController(aiService services.LLM, audioService *services.AudioService) *AudioController {
	return &AudioController{
		aiService:    aiService,
		audioService: audioService,
	}
}

func (ctrl *AudioController) Transcribe(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "audio file is required"})
		return
	}

	audioPath, mimeType, data, err := ctrl.audioService.StoreAndRead(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to process audio file"})
		return
	}

	ctx := c.Request.Context()
	transcript, err := ctrl.aiService.TranscribeAudio(ctx, mimeType, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to transcribe audio"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"transcript": transcript,
		"audio_path": audioPath,
	})
}
