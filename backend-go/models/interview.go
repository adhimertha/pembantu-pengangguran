package models

import "time"

type CompanyType struct {
	Size     string `json:"size"`
	Industry string `json:"industry"`
	Culture  string `json:"culture"`
}

type StartInterviewRequest struct {
	CVText      string      `json:"cv_text" binding:"required"`
	JobSpec     string      `json:"job_spec" binding:"required"`
	CompanyType CompanyType `json:"company_type" binding:"required"`
	UserID      string      `json:"user_id" binding:"required"`
}

type InterviewerPersona struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Style       string `json:"style"`
	AvatarURL   string `json:"avatar_url"`
}

type StartInterviewResponse struct {
	SessionID          string             `json:"session_id"`
	InterviewerPersona InterviewerPersona `json:"interviewer_persona"`
	FirstQuestion      string             `json:"first_question"`
	Status             string             `json:"status"`
}

type RespondRequest struct {
	SessionID  string `json:"session_id" binding:"required"`
	Response   string `json:"response" binding:"required"`
	QuestionID string `json:"question_id" binding:"required"`
}

type RespondResponse struct {
	NextQuestion string                 `json:"next_question"`
	Analysis     map[string]interface{} `json:"analysis"`
	IsComplete   bool                   `json:"is_complete"`
}

type FeedbackRequest struct {
	SessionID string `json:"session_id" binding:"required"`
}

type FeedbackResponse struct {
	OverallScore int                    `json:"overall_score"`
	Breakdown    map[string]interface{} `json:"breakdown"`
	Suggestions  []string               `json:"suggestions"`
	Strengths    []string               `json:"strengths"`
}

type InterviewSession struct {
	ID                 string             `json:"id"`
	UserID             string             `json:"user_id"`
	Status             string             `json:"status"`
	CompanyType        CompanyType        `json:"company_type"`
	JobSpec            string             `json:"job_spec"`
	InterviewerPersona InterviewerPersona `json:"interviewer_persona"`
	StartedAt          time.Time          `json:"started_at"`
	CompletedAt        *time.Time         `json:"completed_at,omitempty"`
}
