import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, RouterLink } from '@angular/router';

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
  feedback: FeedbackData = {
    overallScore: 78,
    categories: [
      {
        name: 'Technical Depth',
        score: 85,
        analysis:
          'You showed strong understanding of system design principles and were able to articulate complex trade-offs clearly.',
      },
      {
        name: 'Communication',
        score: 72,
        analysis:
          'Your explanations are clear, but you could benefit from being more concise in your opening statements.',
      },
      {
        name: 'Cultural Fit',
        score: 77,
        analysis:
          'You align well with the company values, especially regarding ownership and curiosity.',
      },
    ],
    strengths: [
      'Strong problem-solving methodology',
      'Clear articulation of technical trade-offs',
      'Positive and professional attitude',
    ],
    improvements: [
      'Provide more specific examples using the STAR method',
      'Keep behavioral answers under 2 minutes',
      'Ask more insightful questions about the team structure',
    ],
  };

  constructor(private route: ActivatedRoute) {}

  ngOnInit() {
    this.sessionId = this.route.snapshot.paramMap.get('id');
  }
}
