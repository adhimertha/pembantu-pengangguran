import { Routes } from '@angular/router';
import { Home } from './components/home/home';
import { Upload } from './components/upload/upload';
import { Interview } from './components/interview/interview';
import { Feedback } from './components/feedback/feedback';
import { Dashboard } from './components/dashboard/dashboard';

export const routes: Routes = [
  { path: '', redirectTo: 'home', pathMatch: 'full' },
  { path: 'home', component: Home },
  { path: 'upload', component: Upload },
  { path: 'dashboard', component: Dashboard },
  { path: 'interview/:id', component: Interview },
  { path: 'feedback/:id', component: Feedback },
];
