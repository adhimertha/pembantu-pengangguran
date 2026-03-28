package main

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"backend-go/controllers"
	"backend-go/database"
	"backend-go/models"
	"backend-go/routes"
	"backend-go/services"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type fakeLLM struct {
	feedbackCalls int
}

func (f *fakeLLM) GeneratePersona(ctx context.Context, jobSpec string, companyType models.CompanyType) (models.InterviewerPersona, error) {
	return models.InterviewerPersona{
		Name:        "Test Persona",
		Description: "Test Description",
		Style:       "Test Style",
		AvatarURL:   "https://example.com/avatar.png",
	}, nil
}

func (f *fakeLLM) GenerateQuestion(ctx context.Context, persona models.InterviewerPersona, cvText string, history []string) (string, error) {
	if len(history) == 0 {
		return "First question?", nil
	}
	return "Next question?", nil
}

func (f *fakeLLM) AnalyzeResponse(ctx context.Context, question string, response string) (map[string]interface{}, bool, error) {
	return map[string]interface{}{
		"clarity":     0.8,
		"detail":      0.7,
		"relevance":   0.9,
		"is_complete": false,
	}, false, nil
}

func (f *fakeLLM) GenerateFeedback(ctx context.Context, session models.InterviewSession, questions []models.InterviewQuestion, responses []models.InterviewResponse) (models.FeedbackResponse, error) {
	f.feedbackCalls++
	return models.FeedbackResponse{
		OverallScore: 90,
		Breakdown: map[string]interface{}{
			"communication": 88,
			"technical":     92,
			"cultural_fit":  90,
		},
		Suggestions: []string{"Add more metrics.", "Use STAR structure."},
		Strengths:   []string{"Clear communication.", "Strong system thinking."},
	}, nil
}

func (f *fakeLLM) TranscribeAudio(ctx context.Context, mimeType string, audioData []byte) (string, error) {
	return "transcript from audio", nil
}

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	if err := db.AutoMigrate(&models.InterviewSession{}, &models.InterviewQuestion{}, &models.InterviewResponse{}, &models.InterviewFeedback{}); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	database.DB = db
	return db
}

func setupRouter(t *testing.T, uploadDir string, llm *fakeLLM) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	cvService := services.NewCVService(filepath.Join(uploadDir, "cv"), nil, "")
	audioService := services.NewAudioService(filepath.Join(uploadDir, "audio"), 25*1024*1024, nil, "")

	interviewCtrl := controllers.NewInterviewController(llm, audioService)
	cvCtrl := controllers.NewCVController(cvService)
	audioCtrl := controllers.NewAudioController(llm, audioService)

	r := gin.New()
	r.Use(gin.Recovery())
	routes.RegisterRoutes(r, interviewCtrl, cvCtrl, audioCtrl)
	return r
}

func TestE2E_Interview_TextAndAudio(t *testing.T) {
	_ = setupTestDB(t)
	uploadDir := t.TempDir()
	llm := &fakeLLM{}
	r := setupRouter(t, uploadDir, llm)
	srv := httptest.NewServer(r)
	defer srv.Close()

	startReq := models.StartInterviewRequest{
		CVText:  "cv text",
		JobSpec: "job spec",
		CompanyType: models.CompanyType{
			Size:     "Startup",
			Industry: "Fintech",
			Culture:  "Fast-paced",
		},
		UserID: "user_1",
	}
	startBody, _ := json.Marshal(startReq)
	startResp, err := http.Post(srv.URL+"/api/interview/start", "application/json", bytes.NewReader(startBody))
	if err != nil {
		t.Fatalf("start request failed: %v", err)
	}
	defer startResp.Body.Close()
	if startResp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status for start: %d", startResp.StatusCode)
	}
	var startOut models.StartInterviewResponse
	if err := json.NewDecoder(startResp.Body).Decode(&startOut); err != nil {
		t.Fatalf("failed to decode start response: %v", err)
	}
	if startOut.SessionID == "" {
		t.Fatalf("session_id is empty")
	}

	var q1 models.InterviewQuestion
	if err := database.DB.Where("session_id = ?", startOut.SessionID).Order("order_index asc").First(&q1).Error; err != nil {
		t.Fatalf("failed to load first question: %v", err)
	}

	respondReq := models.RespondRequest{
		SessionID:  startOut.SessionID,
		Response:   "my answer",
		QuestionID: q1.ID,
	}
	respondBody, _ := json.Marshal(respondReq)
	respondResp, err := http.Post(srv.URL+"/api/interview/respond", "application/json", bytes.NewReader(respondBody))
	if err != nil {
		t.Fatalf("respond request failed: %v", err)
	}
	defer respondResp.Body.Close()
	if respondResp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status for respond: %d", respondResp.StatusCode)
	}
	var respondOut models.RespondResponse
	if err := json.NewDecoder(respondResp.Body).Decode(&respondOut); err != nil {
		t.Fatalf("failed to decode respond response: %v", err)
	}
	if respondOut.NextQuestion == "" {
		t.Fatalf("next_question is empty")
	}

	var q2 models.InterviewQuestion
	if err := database.DB.Where("session_id = ?", startOut.SessionID).Order("order_index desc").First(&q2).Error; err != nil {
		t.Fatalf("failed to load second question: %v", err)
	}

	var form bytes.Buffer
	writer := multipart.NewWriter(&form)
	_ = writer.WriteField("session_id", startOut.SessionID)
	_ = writer.WriteField("question_id", q2.ID)
	part, err := writer.CreateFormFile("file", "audio.webm")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	_, _ = part.Write([]byte("dummy audio bytes"))
	_ = writer.Close()

	audioReq, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/interview/respond-audio", &form)
	audioReq.Header.Set("Content-Type", writer.FormDataContentType())
	audioResp, err := http.DefaultClient.Do(audioReq)
	if err != nil {
		t.Fatalf("respond-audio request failed: %v", err)
	}
	defer audioResp.Body.Close()
	if audioResp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status for respond-audio: %d", audioResp.StatusCode)
	}
	var audioOut models.RespondResponse
	if err := json.NewDecoder(audioResp.Body).Decode(&audioOut); err != nil {
		t.Fatalf("failed to decode respond-audio response: %v", err)
	}
	if audioOut.Analysis == nil {
		t.Fatalf("analysis is nil")
	}

	feedbackReq := models.FeedbackRequest{SessionID: startOut.SessionID}
	feedbackBody, _ := json.Marshal(feedbackReq)
	feedbackResp, err := http.Post(srv.URL+"/api/interview/feedback", "application/json", bytes.NewReader(feedbackBody))
	if err != nil {
		t.Fatalf("feedback request failed: %v", err)
	}
	defer feedbackResp.Body.Close()
	if feedbackResp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status for feedback: %d", feedbackResp.StatusCode)
	}
	var feedbackOut models.FeedbackResponse
	if err := json.NewDecoder(feedbackResp.Body).Decode(&feedbackOut); err != nil {
		t.Fatalf("failed to decode feedback response: %v", err)
	}
	if feedbackOut.OverallScore == 0 {
		t.Fatalf("overall_score is empty")
	}

	if llm.feedbackCalls != 1 {
		t.Fatalf("expected feedbackCalls=1, got %d", llm.feedbackCalls)
	}

	feedbackResp2, err := http.Post(srv.URL+"/api/interview/feedback", "application/json", bytes.NewReader(feedbackBody))
	if err != nil {
		t.Fatalf("feedback request 2 failed: %v", err)
	}
	defer feedbackResp2.Body.Close()
	if feedbackResp2.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status for feedback 2: %d", feedbackResp2.StatusCode)
	}
	if llm.feedbackCalls != 1 {
		t.Fatalf("expected cached feedback (feedbackCalls still 1), got %d", llm.feedbackCalls)
	}

	req3, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/interview/feedback?force=true", bytes.NewReader(feedbackBody))
	req3.Header.Set("Content-Type", "application/json")
	feedbackResp3, err := http.DefaultClient.Do(req3)
	if err != nil {
		t.Fatalf("feedback request 3 failed: %v", err)
	}
	defer feedbackResp3.Body.Close()
	if feedbackResp3.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status for feedback 3: %d", feedbackResp3.StatusCode)
	}
	if llm.feedbackCalls != 2 {
		t.Fatalf("expected force regenerate (feedbackCalls=2), got %d", llm.feedbackCalls)
	}
}
