import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { ApiClient, CompanyType } from '../../services/api.client';

@Component({
  selector: 'app-upload',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule],
  templateUrl: './upload.html',
  styleUrl: './upload.scss',
})
export class Upload {
  uploadForm: FormGroup;
  isSubmitting = false;
  error: string | null = null;

  constructor(
    private fb: FormBuilder,
    private router: Router,
    private api: ApiClient,
  ) {
    this.uploadForm = this.fb.group({
      cv: [null, Validators.required],
      jobSpec: ['', [Validators.required, Validators.minLength(50)]],
      companyType: ['startup', Validators.required],
      difficulty: ['medium', Validators.required],
    });
  }

  onFileChange(event: any) {
    const file = event.target.files[0];
    if (file) {
      this.uploadForm.patchValue({ cv: file });
    }
  }

  onSubmit() {
    if (!this.uploadForm.valid || this.isSubmitting) return;

    this.error = null;
    this.isSubmitting = true;

    const file = this.uploadForm.get('cv')?.value as File | null;
    const jobSpec = this.uploadForm.get('jobSpec')?.value as string;
    const companyTypeKey = this.uploadForm.get('companyType')?.value as string;
    const userId = this.getOrCreateUserId();

    if (!file) {
      this.error = 'Please select a PDF CV file.';
      this.isSubmitting = false;
      return;
    }

    if (file.type !== 'application/pdf') {
      this.error = 'CV must be a PDF file.';
      this.isSubmitting = false;
      return;
    }

    const companyType = this.mapCompanyType(companyTypeKey);

    this.api.uploadCv(file).subscribe({
      next: (cv) => {
        this.api
          .startInterview({
            cv_text: cv.extracted_text,
            job_spec: jobSpec,
            company_type: companyType,
            user_id: userId,
          })
          .subscribe({
            next: (resp) => {
              this.saveSessionBootstrap(resp.session_id, {
                persona: resp.interviewer_persona,
                companyLabel: companyType.size,
                currentQuestionId: resp.first_question_id,
                currentQuestionText: resp.first_question,
              });
              this.router.navigate(['/interview', resp.session_id]);
              this.isSubmitting = false;
            },
            error: (e: unknown) => {
              this.error = e instanceof Error ? e.message : 'Failed to start interview session';
              this.isSubmitting = false;
            },
          });
      },
      error: (e: unknown) => {
        this.error = e instanceof Error ? e.message : 'Failed to upload CV';
        this.isSubmitting = false;
      },
    });
  }

  private mapCompanyType(key: string): CompanyType {
    const map: Record<string, CompanyType> = {
      startup: {
        size: 'Startup',
        industry: 'General',
        culture: 'Fast-paced, innovative',
      },
      bigtech: {
        size: 'Big Tech',
        industry: 'Technology',
        culture: 'Data-driven, high standards',
      },
      corporate: {
        size: 'Large Corporate',
        industry: 'Enterprise',
        culture: 'Process-driven, risk-aware',
      },
      agency: {
        size: 'Agency',
        industry: 'Consulting',
        culture: 'Client-focused, delivery-oriented',
      },
    };
    return map[key] ?? map['startup'];
  }

  private getOrCreateUserId() {
    const key = 'aiit:user_id';
    if (typeof localStorage === 'undefined') return 'u-anon';
    const existing = localStorage.getItem(key);
    if (existing) return existing;

    const id =
      typeof crypto !== 'undefined' && 'randomUUID' in crypto
        ? `u-${crypto.randomUUID()}`
        : `u-${Date.now()}`;
    localStorage.setItem(key, id);
    return id;
  }

  private saveSessionBootstrap(
    sessionId: string,
    data: {
      persona: unknown;
      companyLabel?: string;
      currentQuestionId: string;
      currentQuestionText: string;
    },
  ) {
    if (typeof localStorage === 'undefined') return;
    const key = `aiit:session:${sessionId}`;
    localStorage.setItem(key, JSON.stringify(data));
  }
}
