package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// CompanyType needs to implement Value and Scan for GORM to handle it as JSONB
type CompanyType struct {
	Size     string `json:"size"`
	Industry string `json:"industry"`
	Culture  string `json:"culture"`
}

func (c CompanyType) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *CompanyType) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, c)
}

// InterviewerPersona needs to implement Value and Scan for GORM to handle it as JSONB
type InterviewerPersona struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Style       string `json:"style"`
	AvatarURL   string `json:"avatar_url"`
}

func (i InterviewerPersona) Value() (driver.Value, error) {
	return json.Marshal(i)
}

func (i *InterviewerPersona) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, i)
}

type StartInterviewRequest struct {
	CVText      string      `json:"cv_text" binding:"required"`
	JobSpec     string      `json:"job_spec" binding:"required"`
	CompanyType CompanyType `json:"company_type" binding:"required"`
	UserID      string      `json:"user_id" binding:"required"`
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

type InterviewFeedback struct {
	ID           string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	SessionID    string    `gorm:"not null;uniqueIndex" json:"session_id"`
	OverallScore int       `json:"overall_score"`
	Breakdown    JSONB     `gorm:"type:jsonb" json:"breakdown"`
	Suggestions  JSONB     `gorm:"type:jsonb" json:"suggestions"`
	Strengths    JSONB     `gorm:"type:jsonb" json:"strengths"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type InterviewSession struct {
	ID                 string             `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	UserID             string             `gorm:"not null" json:"user_id"`
	Status             string             `gorm:"default:'active'" json:"status"`
	CompanyType        CompanyType        `gorm:"type:jsonb" json:"company_type"`
	CVText             string             `gorm:"type:text" json:"cv_text"`
	JobSpec            string             `gorm:"type:text" json:"job_spec"`
	InterviewerPersona InterviewerPersona `gorm:"type:jsonb" json:"interviewer_persona"`
	StartedAt          time.Time          `gorm:"autoCreateTime" json:"started_at"`
	CompletedAt        *time.Time         `json:"completed_at,omitempty"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
	DeletedAt          gorm.DeletedAt     `gorm:"index" json:"-"`
}

type InterviewQuestion struct {
	ID           string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	SessionID    string    `gorm:"not null;index" json:"session_id"`
	QuestionText string    `gorm:"type:text" json:"question_text"`
	Category     string    `json:"category"`
	OrderIndex   int       `json:"order_index"`
	CreatedAt    time.Time `json:"created_at"`
}

type InterviewResponse struct {
	ID           string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	SessionID    string    `gorm:"not null;index" json:"session_id"`
	QuestionID   string    `gorm:"not null;index" json:"question_id"`
	ResponseText string    `gorm:"type:text" json:"response_text"`
	AudioPath    string    `gorm:"type:text" json:"audio_path"`
	Analysis     JSONB     `gorm:"type:jsonb" json:"analysis"`
	CreatedAt    time.Time `json:"created_at"`
}

// Helper type for generic JSONB fields
type JSONB map[string]interface{}

func (j JSONB) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONB) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, j)
}
