package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"backend-go/models"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type AIService struct {
	client *genai.Client
}

func NewAIService(ctx context.Context, apiKey string) (*AIService, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini client: %w", err)
	}
	return &AIService{
		client: client,
	}, nil
}

// GeneratePersona generates a detailed persona based on job spec and company type
func (s *AIService) GeneratePersona(ctx context.Context, jobSpec string, companyType models.CompanyType) (models.InterviewerPersona, error) {
	model := s.client.GenerativeModel("gemini-1.5-flash")
	model.ResponseMIMEType = "application/json"

	prompt := fmt.Sprintf(
		"Based on the following job specification and company details, generate a realistic interviewer persona. "+
			"Job Spec: %s, Company Size: %s, Industry: %s, Culture: %s. "+
			"Return a JSON with fields: name, description, style, avatar_url (use a generic placeholder).",
		jobSpec, companyType.Size, companyType.Industry, companyType.Culture,
	)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return models.InterviewerPersona{}, fmt.Errorf("gemini persona generation failed: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return models.InterviewerPersona{}, fmt.Errorf("no persona generated")
	}

	var persona models.InterviewerPersona
	part := resp.Candidates[0].Content.Parts[0]
	if text, ok := part.(genai.Text); ok {
		if err := json.Unmarshal([]byte(text), &persona); err != nil {
			return models.InterviewerPersona{}, fmt.Errorf("failed to parse persona: %w", err)
		}
	}

	return persona, nil
}

// GenerateQuestion generates the next interview question based on persona, CV, and conversation history
func (s *AIService) GenerateQuestion(ctx context.Context, persona models.InterviewerPersona, cvText string, history []string) (string, error) {
	model := s.client.GenerativeModel("gemini-1.5-flash")

	systemPrompt := fmt.Sprintf(
		"You are %s, described as: %s. Your interview style is %s. "+
			"The candidate's CV says: %s. "+
			"Conduct a realistic interview. Ask one concise, high-impact question at a time.",
		persona.Name, persona.Description, persona.Style, cvText,
	)

	session := model.StartChat()
	session.History = []*genai.Content{
		{
			Role:  "user",
			Parts: []genai.Part{genai.Text(systemPrompt)},
		},
		{
			Role:  "model",
			Parts: []genai.Part{genai.Text("Understood. I am ready to begin the interview.")},
		},
	}

	// Add existing history if any (this is a simplified implementation)
	for _, msg := range history {
		// history should ideally be structured, but for now we'll assume it's just user responses
		session.History = append(session.History, &genai.Content{
			Role:  "user",
			Parts: []genai.Part{genai.Text(msg)},
		})
	}

	resp, err := session.SendMessage(ctx, genai.Text("Ask the next question."))
	if err != nil {
		return "", fmt.Errorf("gemini question generation failed: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no question generated")
	}

	part := resp.Candidates[0].Content.Parts[0]
	if text, ok := part.(genai.Text); ok {
		return strings.TrimSpace(string(text)), nil
	}

	return "", fmt.Errorf("invalid response format")
}

// AnalyzeResponse provides feedback on a single user response
func (s *AIService) AnalyzeResponse(ctx context.Context, question string, response string) (map[string]interface{}, bool, error) {
	model := s.client.GenerativeModel("gemini-1.5-flash")
	model.ResponseMIMEType = "application/json"

	prompt := fmt.Sprintf(
		"Analyze this interview response. Question: %s. Response: %s. "+
			"Return a JSON with fields: 'clarity' (0-1), 'detail' (0-1), 'relevance' (0-1), and 'is_complete' (boolean, true if interview should conclude).",
		question, response,
	)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, false, fmt.Errorf("gemini analysis failed: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, false, fmt.Errorf("no analysis generated")
	}

	var analysis map[string]interface{}
	part := resp.Candidates[0].Content.Parts[0]
	if text, ok := part.(genai.Text); ok {
		if err := json.Unmarshal([]byte(text), &analysis); err != nil {
			return nil, false, fmt.Errorf("failed to parse analysis: %w", err)
		}
	}

	isComplete, _ := analysis["is_complete"].(bool)
	return analysis, isComplete, nil
}

func (s *AIService) GenerateFeedback(ctx context.Context, session models.InterviewSession, questions []models.InterviewQuestion, responses []models.InterviewResponse) (models.FeedbackResponse, error) {
	model := s.client.GenerativeModel("gemini-1.5-flash")
	model.ResponseMIMEType = "application/json"

	type qa struct {
		Question string                 `json:"question"`
		Answer   string                 `json:"answer"`
		Analysis map[string]interface{} `json:"analysis"`
	}

	var pairs []qa
	for i, q := range questions {
		p := qa{Question: q.QuestionText}
		if i < len(responses) {
			p.Answer = responses[i].ResponseText
			p.Analysis = map[string]interface{}(responses[i].Analysis)
		}
		pairs = append(pairs, p)
	}

	input := map[string]interface{}{
		"job_spec": session.JobSpec,
		"company_type": map[string]interface{}{
			"size":     session.CompanyType.Size,
			"industry": session.CompanyType.Industry,
			"culture":  session.CompanyType.Culture,
		},
		"persona": map[string]interface{}{
			"name":        session.InterviewerPersona.Name,
			"description": session.InterviewerPersona.Description,
			"style":       session.InterviewerPersona.Style,
		},
		"qa": pairs,
	}

	inputJSON, err := json.Marshal(input)
	if err != nil {
		return models.FeedbackResponse{}, fmt.Errorf("failed to build feedback input: %w", err)
	}

	prompt := fmt.Sprintf(
		"Given this completed interview session data (job spec, company type, interviewer persona, and Q/A with per-answer analysis), generate final interview feedback.\n"+
			"Return JSON strictly in this schema:\n"+
			"{\"overall_score\": <int 0-100>, \"breakdown\": {\"communication\": <int 0-100>, \"technical\": <int 0-100>, \"cultural_fit\": <int 0-100>}, \"strengths\": [<string>], \"suggestions\": [<string>]}\n"+
			"Be specific and actionable.\n\nSession:\n%s",
		string(inputJSON),
	)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return models.FeedbackResponse{}, fmt.Errorf("gemini feedback generation failed: %w", err)
	}
	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return models.FeedbackResponse{}, fmt.Errorf("no feedback generated")
	}

	var raw string
	if t, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
		raw = string(t)
	} else {
		var parts []string
		for _, p := range resp.Candidates[0].Content.Parts {
			if t, ok := p.(genai.Text); ok {
				parts = append(parts, string(t))
			}
		}
		raw = strings.Join(parts, "\n")
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return models.FeedbackResponse{}, fmt.Errorf("empty feedback")
	}

	var out models.FeedbackResponse
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return models.FeedbackResponse{}, fmt.Errorf("failed to parse feedback: %w", err)
	}

	return out, nil
}

func (s *AIService) TranscribeAudio(ctx context.Context, mimeType string, audioData []byte) (string, error) {
	model := s.client.GenerativeModel("gemini-1.5-flash")

	resp, err := model.GenerateContent(
		ctx,
		genai.Text("Transcribe the provided audio into plain text. Output only the transcript."),
		genai.Blob{MIMEType: mimeType, Data: audioData},
	)
	if err != nil {
		return "", fmt.Errorf("gemini transcription failed: %w", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no transcript generated")
	}

	var parts []string
	for _, p := range resp.Candidates[0].Content.Parts {
		if t, ok := p.(genai.Text); ok {
			parts = append(parts, string(t))
		}
	}
	out := strings.TrimSpace(strings.Join(parts, "\n"))
	if out == "" {
		return "", fmt.Errorf("empty transcript")
	}
	return out, nil
}
