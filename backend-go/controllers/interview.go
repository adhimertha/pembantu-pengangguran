package controllers

import (
	"net/http"
	"strings"
	"time"

	"backend-go/database"
	"backend-go/models"
	"backend-go/services"

	"github.com/gin-gonic/gin"
)

type InterviewController struct {
	aiService    services.LLM
	audioService *services.AudioService
}

func NewInterviewController(aiService services.LLM, audioService *services.AudioService) *InterviewController {
	return &InterviewController{
		aiService:    aiService,
		audioService: audioService,
	}
}

// StartInterview handles POST /api/interview/start
func (ctrl *InterviewController) StartInterview(c *gin.Context) {
	var req models.StartInterviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	// 1. Generate Persona
	persona, err := ctrl.aiService.GeneratePersona(ctx, req.JobSpec, req.CompanyType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate persona"})
		return
	}

	// 2. Generate First Question
	firstQuestion, err := ctrl.aiService.GenerateQuestion(ctx, persona, req.CVText, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate first question"})
		return
	}

	// 3. Save Session to DB
	session := models.InterviewSession{
		UserID:             req.UserID,
		Status:             "active",
		CompanyType:        req.CompanyType,
		CVText:             req.CVText,
		JobSpec:            req.JobSpec,
		InterviewerPersona: persona,
	}

	if err := database.DB.Create(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	// 4. Save First Question
	question := models.InterviewQuestion{
		SessionID:    session.ID,
		QuestionText: firstQuestion,
		OrderIndex:   0,
	}
	database.DB.Create(&question)

	resp := models.StartInterviewResponse{
		SessionID:          session.ID,
		InterviewerPersona: persona,
		FirstQuestion:      firstQuestion,
		Status:             "started",
	}

	c.JSON(http.StatusOK, resp)
}

// RespondToInterview handles POST /api/interview/respond
func (ctrl *InterviewController) RespondToInterview(c *gin.Context) {
	var req models.RespondRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	// 1. Fetch Session and history from DB
	var session models.InterviewSession
	if err := database.DB.First(&session, "id = ?", req.SessionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	// 2. Analyze current response
	var lastQuestion models.InterviewQuestion
	if err := database.DB.First(&lastQuestion, "id = ?", req.QuestionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
		return
	}

	analysis, isComplete, err := ctrl.aiService.AnalyzeResponse(ctx, lastQuestion.QuestionText, req.Response)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to analyze response"})
		return
	}

	// 3. Save Response
	userResponse := models.InterviewResponse{
		SessionID:    req.SessionID,
		QuestionID:   req.QuestionID,
		ResponseText: req.Response,
		AudioPath:    "",
		Analysis:     analysis,
	}
	database.DB.Create(&userResponse)

	// 4. Generate next question if not complete
	var nextQuestionText string
	if !isComplete {
		// Fetch history
		var questions []models.InterviewQuestion
		var responses []models.InterviewResponse
		database.DB.Where("session_id = ?", session.ID).Order("order_index asc").Find(&questions)
		database.DB.Where("session_id = ?", session.ID).Order("created_at asc").Find(&responses)

		// Simple history construction (could be improved)
		var history []string
		for i, q := range questions {
			history = append(history, "Q: "+q.QuestionText)
			if i < len(responses) {
				history = append(history, "A: "+responses[i].ResponseText)
			}
		}

		nextQuestionText, _ = ctrl.aiService.GenerateQuestion(ctx, session.InterviewerPersona, session.CVText, history)

		// Save Next Question
		nextQ := models.InterviewQuestion{
			SessionID:    session.ID,
			QuestionText: nextQuestionText,
			OrderIndex:   len(questions),
		}
		database.DB.Create(&nextQ)
	} else {
		// Mark session as completed
		now := time.Now()
		database.DB.Model(&session).Updates(models.InterviewSession{Status: "completed", CompletedAt: &now})
	}

	resp := models.RespondResponse{
		NextQuestion: nextQuestionText,
		Analysis:     analysis,
		IsComplete:   isComplete,
	}

	c.JSON(http.StatusOK, resp)
}

func (ctrl *InterviewController) RespondToInterviewAudio(c *gin.Context) {
	sessionID := c.PostForm("session_id")
	questionID := c.PostForm("question_id")
	if sessionID == "" || questionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id and question_id are required"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "audio file is required"})
		return
	}

	if ctrl.audioService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "audio service is not configured"})
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

	var session models.InterviewSession
	if err := database.DB.First(&session, "id = ?", sessionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	var lastQuestion models.InterviewQuestion
	if err := database.DB.First(&lastQuestion, "id = ?", questionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
		return
	}

	analysis, isComplete, err := ctrl.aiService.AnalyzeResponse(ctx, lastQuestion.QuestionText, transcript)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to analyze response"})
		return
	}

	userResponse := models.InterviewResponse{
		SessionID:    sessionID,
		QuestionID:   questionID,
		ResponseText: transcript,
		AudioPath:    audioPath,
		Analysis:     analysis,
	}
	database.DB.Create(&userResponse)

	var nextQuestionText string
	if !isComplete {
		var questions []models.InterviewQuestion
		var responses []models.InterviewResponse
		database.DB.Where("session_id = ?", session.ID).Order("order_index asc").Find(&questions)
		database.DB.Where("session_id = ?", session.ID).Order("created_at asc").Find(&responses)

		var history []string
		for i, q := range questions {
			history = append(history, "Q: "+q.QuestionText)
			if i < len(responses) {
				history = append(history, "A: "+responses[i].ResponseText)
			}
		}

		nextQuestionText, _ = ctrl.aiService.GenerateQuestion(ctx, session.InterviewerPersona, session.CVText, history)

		nextQ := models.InterviewQuestion{
			SessionID:    session.ID,
			QuestionText: nextQuestionText,
			OrderIndex:   len(questions),
		}
		database.DB.Create(&nextQ)
	} else {
		now := time.Now()
		database.DB.Model(&session).Updates(models.InterviewSession{Status: "completed", CompletedAt: &now})
	}

	resp := models.RespondResponse{
		NextQuestion: nextQuestionText,
		Analysis: map[string]interface{}{
			"transcript": transcript,
			"analysis":   analysis,
			"audio_path": audioPath,
		},
		IsComplete: isComplete,
	}
	c.JSON(http.StatusOK, resp)
}

// GenerateFeedback handles POST /api/interview/feedback
func (ctrl *InterviewController) GenerateFeedback(c *gin.Context) {
	var req models.FeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	force := isTruthy(c.Query("force"))
	if !force {
		var cached models.InterviewFeedback
		if err := database.DB.First(&cached, "session_id = ?", req.SessionID).Error; err == nil {
			c.JSON(http.StatusOK, feedbackRecordToResponse(cached))
			return
		}
	}

	ctx := c.Request.Context()

	var session models.InterviewSession
	if err := database.DB.First(&session, "id = ?", req.SessionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	var questions []models.InterviewQuestion
	var responses []models.InterviewResponse
	database.DB.Where("session_id = ?", session.ID).Order("order_index asc").Find(&questions)
	database.DB.Where("session_id = ?", session.ID).Order("created_at asc").Find(&responses)

	if len(questions) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No interview questions found for this session"})
		return
	}

	feedback, err := ctrl.aiService.GenerateFeedback(ctx, session, questions, responses)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate feedback"})
		return
	}

	var existing models.InterviewFeedback
	if err := database.DB.First(&existing, "session_id = ?", session.ID).Error; err == nil {
		existing.OverallScore = feedback.OverallScore
		existing.Breakdown = models.JSONB(feedback.Breakdown)
		existing.Suggestions = models.JSONB{"items": feedback.Suggestions}
		existing.Strengths = models.JSONB{"items": feedback.Strengths}
		if err := database.DB.Save(&existing).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to persist feedback"})
			return
		}
	} else {
		record := models.InterviewFeedback{
			SessionID:    session.ID,
			OverallScore: feedback.OverallScore,
			Breakdown:    models.JSONB(feedback.Breakdown),
			Suggestions:  models.JSONB{"items": feedback.Suggestions},
			Strengths:    models.JSONB{"items": feedback.Strengths},
		}
		if err := database.DB.Create(&record).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to persist feedback"})
			return
		}
	}

	c.JSON(http.StatusOK, feedback)
}

func isTruthy(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}

func feedbackRecordToResponse(f models.InterviewFeedback) models.FeedbackResponse {
	return models.FeedbackResponse{
		OverallScore: f.OverallScore,
		Breakdown:    map[string]interface{}(f.Breakdown),
		Suggestions:  jsonbToStringSlice(f.Suggestions),
		Strengths:    jsonbToStringSlice(f.Strengths),
	}
}

func jsonbToStringSlice(v models.JSONB) []string {
	if v == nil {
		return nil
	}
	if items, ok := v["items"]; ok {
		return anyToStringSlice(items)
	}
	return anyToStringSlice(map[string]interface{}(v))
}

func anyToStringSlice(v interface{}) []string {
	switch x := v.(type) {
	case []string:
		return x
	case []interface{}:
		out := make([]string, 0, len(x))
		for _, it := range x {
			if s, ok := it.(string); ok {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}
