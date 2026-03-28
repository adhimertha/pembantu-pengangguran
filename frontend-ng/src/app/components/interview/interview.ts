import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';

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
  persona = {
    name: 'Sarah',
    role: 'Senior Technical Lead',
    company: 'Big Tech',
    avatar: '👩‍💻',
  };

  constructor(
    private route: ActivatedRoute,
    private router: Router,
  ) {}

  ngOnInit() {
    this.sessionId = this.route.snapshot.paramMap.get('id');
    this.startInterview();
  }

  startInterview() {
    this.isTyping = true;
    setTimeout(() => {
      this.messages.push({
        id: '1',
        text: `Hello! I'm ${this.persona.name}, a ${this.persona.role} here. I've reviewed your CV and the job description. Ready to start?`,
        sender: 'ai',
        timestamp: new Date(),
      });
      this.isTyping = false;

      // First actual question
      setTimeout(() => {
        this.addAiMessage(
          "Great. Let's dive in. Can you tell me about a complex technical challenge you faced recently and how you solved it?",
        );
      }, 1000);
    }, 1500);
  }

  sendMessage() {
    if (!this.userInput.trim()) return;

    const userMsg: Message = {
      id: Date.now().toString(),
      text: this.userInput,
      sender: 'user',
      timestamp: new Date(),
    };
    this.messages.push(userMsg);
    this.userInput = '';

    // Simulate AI response
    this.isTyping = true;
    setTimeout(() => {
      this.addAiMessage(
        "That's an interesting approach. How did you handle the scalability aspects of that solution?",
      );
      this.isTyping = false;
    }, 2000);
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
}
