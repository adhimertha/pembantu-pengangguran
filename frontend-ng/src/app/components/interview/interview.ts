import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { ApiClient, InterviewerPersona } from '../../services/api.client';

interface Message {
  id: string;
  text: string;
  sender: 'ai' | 'user';
  timestamp: Date;
}

@Component({
  selector: 'app-interview',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './interview.html',
  styleUrl: './interview.scss',
})
export class Interview implements OnInit {
  sessionId: string | null = null;
  messages: Message[] = [];
  userInput: string = '';
  isTyping: boolean = false;
  error: string | null = null;
  persona: InterviewerPersona = {
    name: 'Interviewer',
    description: '',
    style: '',
    avatar_url: '',
  };
  companyLabel = '';
  currentQuestionId: string | null = null;
  currentQuestionText: string | null = null;

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private api: ApiClient,
  ) {}

  ngOnInit() {
    this.sessionId = this.route.snapshot.paramMap.get('id');
    if (!this.sessionId) {
      this.error = 'Missing session id.';
      return;
    }

    const bootstrap = this.loadSessionBootstrap(this.sessionId);
    if (!bootstrap) {
      this.error = 'Session data not found. Please start from Upload.';
      return;
    }

    this.persona = bootstrap.persona;
    this.companyLabel = bootstrap.companyLabel ?? '';
    this.currentQuestionId = bootstrap.currentQuestionId;
    this.currentQuestionText = bootstrap.currentQuestionText;

    this.messages = [];
    this.addAiMessage(
      `Hi, I'm ${this.persona.name}. ${this.persona.style ? this.persona.style + '.' : ''} Ready to begin?`,
    );
    if (this.currentQuestionText) {
      this.addAiMessage(this.currentQuestionText);
    }
  }

  sendMessage() {
    if (!this.userInput.trim() || this.isTyping) return;
    if (!this.sessionId || !this.currentQuestionId) return;
    this.error = null;

    const userMsg: Message = {
      id: Date.now().toString(),
      text: this.userInput,
      sender: 'user',
      timestamp: new Date(),
    };
    this.messages.push(userMsg);
    this.userInput = '';

    this.isTyping = true;
    this.api
      .respond({
        session_id: this.sessionId,
        response: userMsg.text,
        question_id: this.currentQuestionId,
      })
      .subscribe({
        next: (resp) => {
          this.isTyping = false;
          if (resp.is_complete) {
            this.router.navigate(['/feedback', this.sessionId]);
            return;
          }

          this.currentQuestionId = resp.next_question_id;
          this.currentQuestionText = resp.next_question;
          this.saveSessionBootstrap(this.sessionId!, {
            persona: this.persona,
            companyLabel: this.companyLabel,
            currentQuestionId: this.currentQuestionId!,
            currentQuestionText: this.currentQuestionText ?? '',
          });
          if (resp.next_question) {
            this.addAiMessage(resp.next_question);
          }
        },
        error: (e: unknown) => {
          this.isTyping = false;
          this.error = e instanceof Error ? e.message : 'Failed to send response';
        },
      });
  }

  addAiMessage(text: string) {
    this.messages.push({
      id: Date.now().toString(),
      text: text,
      sender: 'ai',
      timestamp: new Date(),
    });
  }

  endInterview() {
    if (
      confirm(
        'Are you sure you want to end the interview? You will receive feedback based on your responses so far.',
      )
    ) {
      this.router.navigate(['/feedback', this.sessionId]);
    }
  }

  private loadSessionBootstrap(sessionId: string): {
    persona: InterviewerPersona;
    companyLabel?: string;
    currentQuestionId: string;
    currentQuestionText: string;
  } | null {
    try {
      if (typeof localStorage === 'undefined') return null;
      const raw = localStorage.getItem(`aiit:session:${sessionId}`);
      if (!raw) return null;
      const parsed = JSON.parse(raw) as any;
      if (!parsed?.persona || !parsed?.currentQuestionId) return null;
      return parsed;
    } catch {
      return null;
    }
  }

  private saveSessionBootstrap(
    sessionId: string,
    data: {
      persona: InterviewerPersona;
      companyLabel?: string;
      currentQuestionId: string;
      currentQuestionText: string;
    },
  ) {
    if (typeof localStorage === 'undefined') return;
    localStorage.setItem(`aiit:session:${sessionId}`, JSON.stringify(data));
  }
}
