# Pengangguran API Contract (v1)

This document outlines the communication contract between the Angular frontend and Go backend.

---

## **Base URL**
`http://localhost:8080/api`

---

## **1. CV Management**

### **1.1 Upload CV**
Uploads a CV in PDF format and returns the extracted text. This text should then be used in the `Start Interview Session` request.

*   **URL**: `/cv/upload`
*   **Method**: `POST`
*   **Content-Type**: `multipart/form-data`
*   **Authentication Required**: No (currently)

#### **Form Parameters**
| Parameter | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `file` | `file` | Yes | The CV file (must be PDF). |

#### **Success Response (200 OK)**
| Field | Type | Description |
| :--- | :--- | :--- |
| `filename` | `string` | Name of the uploaded file. |
| `extracted_text` | `string` | The plain text content extracted from the PDF. |
| `message` | `string` | Success message. |

**Example Response:**
```json
{
  "filename": "cv_alex_dev.pdf",
  "extracted_text": "Experienced Go Developer with 5 years in fintech...",
  "message": "CV uploaded and parsed successfully"
}
```

---

## **2. Interview Session Management**

### **2.1 Start Interview Session**
Initiates a new interview session. This endpoint extracts the user's CV details, analyzes the job specification, generates a specific interviewer persona, and returns the first question.

*   **URL**: `/interview/start`
*   **Method**: `POST`
*   **Authentication Required**: No

#### **Request Body**
| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `cv_text` | `string` | Yes | Extracted text from the user's CV (PDF). |
| `job_spec` | `string` | Yes | Text describing the job role, requirements, and responsibilities. |
| `company_type` | `object` | Yes | An object containing `size`, `industry`, and `culture`. |
| `user_id` | `string` | Yes | Unique identifier for the job seeker. |

**Example Request:**
```json
{
  "cv_text": "Experienced Go Developer with 5 years in fintech...",
  "job_spec": "Looking for a Senior Backend Engineer to lead our infrastructure team...",
  "company_type": {
    "size": "Startup",
    "industry": "Fintech",
    "culture": "Fast-paced, innovative"
  },
  "user_id": "user_12345"
}
```

#### **Success Response (200 OK)**
| Field | Type | Description |
| :--- | :--- | :--- |
| `session_id` | `string` | Unique identifier for the interview session. |
| `interviewer_persona` | `object` | Details of the AI-generated interviewer. |
| `first_question` | `string` | The initial question from the AI interviewer. |
| `status` | `string` | Status of the session (e.g., "started"). |

**Example Response:**
```json
{
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "interviewer_persona": {
    "name": "Sarah the CTO",
    "description": "A pragmatic and deeply technical leader with 15 years in tech.",
    "style": "Direct, focuses on scalability and architectural decisions.",
    "avatar_url": "https://example.com/avatars/sarah.png"
  },
  "first_question": "Based on your experience at Fintech Corp, how would you design a rate-limiting service for our high-traffic API?",
  "status": "started"
}
```

---

### **2.2 Respond to Interview Question**
Sends the user's response to the current question and receives the next follow-up question or ends the session.

*   **URL**: `/interview/respond`
*   **Method**: `POST`
*   **Authentication Required**: No

#### **Request Body**
| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `session_id` | `string` | Yes | The ID of the current interview session. |
| `response` | `string` | Yes | The candidate's answer to the previous question. |
| `question_id` | `string` | Yes | The ID of the question being answered. |

**Example Request:**
```json
{
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "response": "I would use a distributed Token Bucket algorithm implemented with Redis...",
  "question_id": "q_1"
}
```

#### **Success Response (200 OK)**
| Field | Type | Description |
| :--- | :--- | :--- |
| `next_question` | `string` | The follow-up question from the AI (null if session complete). |
| `analysis` | `object` | Real-time feedback/scoring of the response. |
| `is_complete` | `boolean` | True if the interview has concluded. |

**Example Response:**
```json
{
  "next_question": "Excellent. And how would you ensure data consistency between Redis and your primary database in that design?",
  "analysis": {
    "clarity": 0.9,
    "detail": 0.85,
    "relevance": 0.95
  },
  "is_complete": false
}
```

---

### **2.3 Respond to Interview With Audio**
Uploads a recorded answer, transcribes it using Gemini, saves it, and continues the interview like `/interview/respond`.

*   **URL**: `/interview/respond-audio`
*   **Method**: `POST`
*   **Content-Type**: `multipart/form-data`
*   **Authentication Required**: No

#### **Form Parameters**
| Parameter | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `session_id` | `string` | Yes | The ID of the current interview session. |
| `question_id` | `string` | Yes | The ID of the question being answered. |
| `file` | `file` | Yes | Recorded audio file (webm/wav/mpeg supported; depends on browser recorder). |

#### **Success Response (200 OK)**
| Field | Type | Description |
| :--- | :--- | :--- |
| `next_question` | `string` | The follow-up question from the AI (empty if session complete). |
| `analysis` | `object` | Contains `transcript`, `analysis`, and `audio_path`. |
| `is_complete` | `boolean` | True if the interview has concluded. |

**Example Response:**
```json
{
  "next_question": "Thanks. Can you explain how you would monitor and alert on this system in production?",
  "analysis": {
    "transcript": "I would use Redis with a token bucket approach ...",
    "analysis": { "clarity": 0.86, "detail": 0.74, "relevance": 0.9, "is_complete": false },
    "audio_path": "./uploads/audio/2c2a79a8-acde-4f3a-acde-2d5b3a4c9c0f.webm"
  },
  "is_complete": false
}
```

---

### **2.4 Generate Final Feedback**
Requests a comprehensive summary and score for a completed interview session.

*   **URL**: `/interview/feedback`
*   **Method**: `POST`
*   **Authentication Required**: No
*   **Caching**: If feedback already exists for the session, the backend returns the cached feedback by default.
*   **Force Regenerate**: Add query param `?force=true` to re-run Gemini and overwrite cached feedback.

#### **Request Body**
| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `session_id` | `string` | Yes | The ID of the completed interview session. |

**Example Request:**
```json
{
  "session_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

#### **Success Response (200 OK)**
| Field | Type | Description |
| :--- | :--- | :--- |
| `overall_score` | `number` | A total score out of 100. |
| `breakdown` | `object` | Scores for specific categories (e.g., technical, cultural). |
| `suggestions` | `array` | A list of strings for improvement. |
| `strengths` | `array` | A list of strings identifying strong performance areas. |

**Example Response:**
```json
{
  "overall_score": 88,
  "breakdown": {
    "communication": 90,
    "technical": 85,
    "cultural_fit": 92
  },
  "suggestions": [
    "Provide more concrete metrics when discussing past project successes.",
    "Practice explaining complex distributed system concepts to non-technical stakeholders."
  ],
  "strengths": [
    "Deep understanding of Go concurrency primitives.",
    "Very clear communication style.",
    "Strong alignment with startup culture."
  ]
}
```

---

## **3. Audio Utilities**

### **3.1 Transcribe Audio**
Uploads an audio file and returns a transcript (useful if frontend wants to call `/interview/respond` with text instead of using `/interview/respond-audio`).

*   **URL**: `/audio/transcribe`
*   **Method**: `POST`
*   **Content-Type**: `multipart/form-data`
*   **Authentication Required**: No (currently)

#### **Form Parameters**
| Parameter | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `file` | `file` | Yes | Recorded audio file. |

#### **Success Response (200 OK)**
| Field | Type | Description |
| :--- | :--- | :--- |
| `transcript` | `string` | Extracted transcript text. |
| `audio_path` | `string` | Stored local file path (temporary implementation). |

**Example Response:**
```json
{
  "transcript": "I would use Redis with a token bucket approach ...",
  "audio_path": "./uploads/audio/2c2a79a8-acde-4f3a-acde-2d5b3a4c9c0f.webm"
}
```

---

## **Error Responses**
All error responses will return a non-2xx status code and follow this format:

```json
{
  "error": "A descriptive error message explaining what went wrong."
}
```

| Status Code | Description |
| :--- | :--- |
| `400 Bad Request` | Invalid request body or missing required fields. |
| `401 Unauthorized` | Missing or invalid authentication token. |
| `404 Not Found` | The requested session or resource does not exist. |
| `500 Internal Server Error` | Unexpected server error (e.g., AI service failure). |
