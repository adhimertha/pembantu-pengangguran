import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';

type StoredSession = {
  id: string;
  createdAt: string;
  overallScore: number;
};

const STORAGE_KEY = 'aiit:sessions';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './dashboard.html',
  styleUrl: './dashboard.scss',
})
export class Dashboard implements OnInit {
  sessions: StoredSession[] = [];

  ngOnInit() {
    this.sessions = this.loadSessions();
  }

  get hasSessions() {
    return this.sessions.length > 0;
  }

  get bestScore() {
    return this.sessions.reduce((acc, s) => Math.max(acc, s.overallScore), 0);
  }

  get averageScore() {
    if (!this.hasSessions) return 0;
    const sum = this.sessions.reduce((acc, s) => acc + s.overallScore, 0);
    return Math.round(sum / this.sessions.length);
  }

  get lastSession() {
    return this.sessions[0] ?? null;
  }

  get chartPoints() {
    const last = this.sessions.slice(0, 8).reverse();
    if (last.length === 0) return '';

    const width = 320;
    const height = 110;
    const padding = 10;

    const xs = last.map((_, i) => {
      if (last.length === 1) return width / 2;
      return padding + (i * (width - padding * 2)) / (last.length - 1);
    });

    const ys = last.map((s) => {
      const v = Math.max(0, Math.min(100, s.overallScore));
      const t = v / 100;
      return padding + (1 - t) * (height - padding * 2);
    });

    return xs.map((x, i) => `${x.toFixed(1)},${ys[i].toFixed(1)}`).join(' ');
  }

  private loadSessions(): StoredSession[] {
    try {
      const raw = localStorage.getItem(STORAGE_KEY);
      if (!raw) return [];
      const parsed = JSON.parse(raw) as unknown;
      if (!Array.isArray(parsed)) return [];

      const normalized = parsed
        .map((x) => {
          const obj = x as Partial<StoredSession>;
          if (!obj.id || !obj.createdAt || typeof obj.overallScore !== 'number') return null;
          return {
            id: String(obj.id),
            createdAt: String(obj.createdAt),
            overallScore: Math.max(0, Math.min(100, Number(obj.overallScore))),
          } satisfies StoredSession;
        })
        .filter((x): x is StoredSession => x !== null)
        .sort((a, b) => b.createdAt.localeCompare(a.createdAt));

      return normalized;
    } catch {
      return [];
    }
  }
}

