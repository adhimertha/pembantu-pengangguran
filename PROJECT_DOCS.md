# Interview Training Platform (Project: Pengangguran) 🚀

This document serves as our shared "brain" and collaboration guide. It outlines the architecture, vision, and how we work together to build this AI-powered interview trainer.

---

## 🛠 Tech Stack
- **Frontend**: [Angular v21](file:///Users/adhimertha/Documents/trae_projects/pengangguran/frontend-ng) (Standalone, SSR, SCSS)
- **Backend**: [Go (Golang)](file:///Users/adhimertha/Documents/trae_projects/pengangguran/backend-go) (Gin, OpenAI SDK, PDF Parsing)
- **AI**: LLM-based interviewer with company-specific personas.

---

## 🎯 Project Vision
To help job seekers (especially in Indonesia - *Pengangguran*) prepare for interviews by simulating real scenarios using their own CV and the specific job description they are applying for.

### **Core Workflow**
1. **Input**: User uploads CV (PDF) + pastes Job Description + selects Company Type.
2. **Analysis**: AI extracts skills from CV and matches them against the JD.
3. **Simulation**: AI acts as an interviewer (Startup, Corporate, or Big Tech).
4. **Feedback**: AI provides a "Confidence Score" and areas for improvement.

---

## 📂 Project Structure
```text
/pengangguran
  ├── backend-go/      # Go source code (API, AI Logic)
  ├── frontend-ng/     # Angular v21 source code (UI/UX)
  ├── docs/            # Additional documentation & assets
  └── .gitignore       # Root-level ignore rules
```

---

## 🤝 How We Work Together (Team Guidelines)

### **1. AI's Role (Trae AI)**
- I handle the heavy lifting: scaffolding, implementing complex logic, and managing dependencies.
- I will proactively suggest architecture improvements and security best practices.
- I'll maintain this document as the project evolves.

### **2. Your Role (The Lead Developer)**
- Provide the vision and specific requirements.
- Review and approve the code changes I propose.
- Test the application in your local environment.

### **3. Communication Protocol**
- **Clarity**: We use clear, actionable tasks.
- **Verification**: After every major change, we verify the build and functionality.
- **Iteration**: We build in phases (MVP first, then advanced features).

---

## 📍 Current Roadmap

- [x] Project Scaffolding (Go & Angular)
- [x] Root .gitignore & Project Documentation
- [ ] **Next**: Define Backend API Endpoints (CV Upload & Chat)
- [ ] **Next**: Implement AI Prompt Engineering (Interview Personas)
- [ ] **Next**: Build Angular Dashboard & Chat Interface
- [ ] **Future**: Voice integration (STT/TTS)

---

*Last Updated: 2026-03-28*
