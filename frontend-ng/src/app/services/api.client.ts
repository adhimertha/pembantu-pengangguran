import { inject, Injectable } from '@angular/core';
import { HttpClient, HttpErrorResponse } from '@angular/common/http';
import { catchError, Observable, throwError } from 'rxjs';
import { API_BASE_URL } from './api.config';

export type UploadCvResponse = {
  filename: string;
  extracted_text: string;
  message: string;
};

export type CompanyType = {
  size: string;
  industry: string;
  culture: string;
};

export type StartInterviewRequest = {
  cv_text: string;
  job_spec: string;
  company_type: CompanyType;
  user_id: string;
};

export type InterviewerPersona = {
  name: string;
  description: string;
  style: string;
  avatar_url: string;
};

export type StartInterviewResponse = {
  session_id: string;
  interviewer_persona: InterviewerPersona;
  first_question: string;
  first_question_id: string;
  status: string;
};

export type RespondRequest = {
  session_id: string;
  response: string;
  question_id: string;
};

export type RespondResponse = {
  next_question: string;
  next_question_id: string;
  analysis: Record<string, unknown>;
  is_complete: boolean;
};

export type FeedbackRequest = {
  session_id: string;
};

export type FeedbackResponse = {
  overall_score: number;
  breakdown: Record<string, number>;
  suggestions: string[];
  strengths: string[];
};

@Injectable({ providedIn: 'root' })
export class ApiClient {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = inject(API_BASE_URL);

  uploadCv(file: File): Observable<UploadCvResponse> {
    const form = new FormData();
    form.append('file', file);
    return this.http
      .post<UploadCvResponse>(this.url('/cv/upload'), form)
      .pipe(catchError((e) => this.handleError(e)));
  }

  startInterview(payload: StartInterviewRequest): Observable<StartInterviewResponse> {
    return this.http
      .post<StartInterviewResponse>(this.url('/interview/start'), payload)
      .pipe(catchError((e) => this.handleError(e)));
  }

  respond(payload: RespondRequest): Observable<RespondResponse> {
    return this.http
      .post<RespondResponse>(this.url('/interview/respond'), payload)
      .pipe(catchError((e) => this.handleError(e)));
  }

  respondAudio(sessionId: string, questionId: string, file: File): Observable<RespondResponse> {
    const form = new FormData();
    form.append('session_id', sessionId);
    form.append('question_id', questionId);
    form.append('file', file);
    return this.http
      .post<RespondResponse>(this.url('/interview/respond-audio'), form)
      .pipe(catchError((e) => this.handleError(e)));
  }

  generateFeedback(payload: FeedbackRequest, force = false): Observable<FeedbackResponse> {
    const url = force ? this.url('/interview/feedback?force=true') : this.url('/interview/feedback');
    return this.http
      .post<FeedbackResponse>(url, payload)
      .pipe(catchError((e) => this.handleError(e)));
  }

  private url(path: string) {
    return `${this.baseUrl}${path}`;
  }

  private handleError(err: unknown) {
    if (err instanceof HttpErrorResponse) {
      const maybe = err.error as { error?: unknown } | null;
      const msg =
        typeof maybe?.error === 'string'
          ? maybe.error
          : typeof err.message === 'string'
            ? err.message
            : 'Request failed';
      return throwError(() => new Error(msg));
    }
    return throwError(() => new Error('Request failed'));
  }
}

