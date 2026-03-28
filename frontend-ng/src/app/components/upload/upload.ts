import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule } from '@angular/forms';
import { Router } from '@angular/router';

@Component({
  selector: 'app-upload',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule],
  templateUrl: './upload.html',
  styleUrl: './upload.scss',
})
export class Upload {
  uploadForm: FormGroup;

  constructor(
    private fb: FormBuilder,
    private router: Router,
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
    if (this.uploadForm.valid) {
      const sessionId = `s-${Date.now()}`;
      this.router.navigate(['/interview', sessionId]);
    }
  }
}
