package services

import (
	"context"

	"backend-go/models"
)

type LLM interface {
	GeneratePersona(ctx context.Context, jobSpec string, companyType models.CompanyType) (models.InterviewerPersona, error)
	GenerateQuestion(ctx context.Context, persona models.InterviewerPersona, cvText string, history []string) (string, error)
	AnalyzeResponse(ctx context.Context, question string, response string) (map[string]interface{}, bool, error)
	GenerateFeedback(ctx context.Context, session models.InterviewSession, questions []models.InterviewQuestion, responses []models.InterviewResponse) (models.FeedbackResponse, error)
	TranscribeAudio(ctx context.Context, mimeType string, audioData []byte) (string, error)
}
