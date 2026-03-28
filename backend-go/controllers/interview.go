package controllers

import (
	"net/http"

	"backend-go/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// StartInterview handles POST /api/interview/start
func StartInterview(c *gin.Context) {
	var req models.StartInterviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// For now, return mock data
	sessionID := uuid.New().String()
	persona := models.InterviewerPersona{
		Name:        "Alex the Analyst",
		Description: "A sharp and thorough technical lead from a top tech company.",
		Style:       "Direct, technical, and detail-oriented",
		AvatarURL:   "https://example.com/avatar/alex.png",
	}

	resp := models.StartInterviewResponse{
		SessionID:          sessionID,
		InterviewerPersona: persona,
		FirstQuestion:      "Could you please walk me through a challenging project you've worked on recently and the technical hurdles you faced?",
		Status:             "started",
	}

	c.JSON(http.StatusOK, resp)
}

// RespondToInterview handles POST /api/interview/respond
func RespondToInterview(c *gin.Context) {
	var req models.RespondRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// For now, return mock data
	resp := models.RespondResponse{
		NextQuestion: "That's interesting. How did you handle the trade-offs between performance and scalability in that specific scenario?",
		Analysis: map[string]interface{}{
			"clarity": 0.85,
			"detail":  0.7,
		},
		IsComplete: false,
	}

	c.JSON(http.StatusOK, resp)
}

// GenerateFeedback handles POST /api/interview/feedback
func GenerateFeedback(c *gin.Context) {
	var req models.FeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// For now, return mock data
	resp := models.FeedbackResponse{
		OverallScore: 82,
		Breakdown: map[string]interface{}{
			"communication": 85,
			"technical":     78,
			"cultural":      88,
		},
		Suggestions: []string{
			"Try to provide more concrete examples of when you used specific technologies.",
			"Structure your answers using the STAR method (Situation, Task, Action, Result).",
		},
		Strengths: []string{
			"Strong understanding of system design principles.",
			"Excellent communication skills and clarity of thought.",
		},
	}

	c.JSON(http.StatusOK, resp)
}
