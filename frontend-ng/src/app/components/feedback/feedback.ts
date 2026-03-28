import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { ApiClient, FeedbackResponse } from '../../services/api.client';

interface FeedbackData {
  overallScore: number;
  categories: {
    name: string;
    score: number;
    analysis: string;
  }[];
  strengths: string[];
  improvements: string[];
}

type StoredSession = {
  id: string;
  createdAt: string;
  overallScore: number;
};

const STORAGE_KEY = 'aiit:sessions';

@Component({
  selector: 'app-feedback',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './feedback.html',
  styleUrl: './feedback.scss',
})
export class Feedback implements OnInit {
  sessionId: string | null = null;
  today: Date = new Date();
  isLoading = true;
  error: string | null = null;
  feedback: FeedbackData = {
    overallScore: 0,
    categories: [],
    strengths: [],
    improvements: [],
  };

  constructor(
    private route: ActivatedRoute,
    private api: ApiClient,
  ) {}

  ngOnInit() {
    this.sessionId = this.route.snapshot.paramMap.get('id');
    if (!this.sessionId) {
      this.error = 'Missing session id.';
      this.isLoading = false;
      return;
    }

    this.isLoading = true;
    this.error = null;
    this.api.generateFeedback({ session_id: this.sessionId }).subscribe({
      next: (resp) => {
        this.feedback = this.mapFeedback(resp);
        this.saveSession({
          id: this.sessionId!,
          createdAt: new Date().toISOString(),
          overallScore: this.feedback.overallScore,
        });
        this.isLoading = false;
      },
      error: (e: unknown) => {
        this.error = e instanceof Error ? e.message : 'Failed to load feedback';
        this.isLoading = false;
      },
    });
  }

  private mapFeedback(resp: FeedbackResponse): FeedbackData {
    const breakdown = resp.breakdown ?? {};
    const categories = Object.entries(breakdown).map(([k, v]) => {
      const name = this.humanizeBreakdownKey(k);
      const score = typeof v === 'number' ? v : Number(v);
      return { name, score, analysis: '' };
    });

    return {
      overallScore: resp.overall_score ?? 0,
      categories,
      strengths: resp.strengths ?? [],
      improvements: resp.suggestions ?? [],
    };
  }

  private humanizeBreakdownKey(key: string) {
    const map: Record<string, string> = {
      communication: 'Communication',
      technical: 'Technical',
      cultural_fit: 'Cultural Fit',
    };
    return map[key] ?? key.replace(/_/g, ' ');
  }

  private saveSession(session: StoredSession) {
    try {
      if (typeof localStorage === 'undefined') return;
      const raw = localStorage.getItem(STORAGE_KEY);
      const existing = raw ? (JSON.parse(raw) as unknown) : [];
      const list = Array.isArray(existing) ? (existing as StoredSession[]) : [];

      const normalized: StoredSession = {
        id: String(session.id),
        createdAt: String(session.createdAt),
        overallScore: Math.max(0, Math.min(100, Number(session.overallScore))),
      };

      const withoutSameId = list.filter((s) => (s as StoredSession).id !== normalized.id);
      const next = [normalized, ...withoutSameId].slice(0, 50);
      localStorage.setItem(STORAGE_KEY, JSON.stringify(next));
    } catch {}
  }
}
