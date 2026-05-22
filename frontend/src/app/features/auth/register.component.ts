import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { TranslateModule } from '@ngx-translate/core';
import { AuthService } from '../../core/services/auth.service';

@Component({
  selector: 'app-register',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterLink, TranslateModule],
  template: `
    <div class="auth-page">
      <div class="auth-bg">
        <div class="float-elem f1"></div>
        <div class="float-elem f2"></div>
        <div class="float-elem f3">🐹</div>
        <div class="float-elem f4"></div>
      </div>
      <div class="auth-card card slide-up">
        <div class="auth-logo"></div>
        <h1 class="auth-title">{{ 'AUTH.REG_TITLE' | translate }}</h1>
        <p class="auth-sub">{{ 'AUTH.REG_SUB' | translate }}</p>

        <div class="alert alert-error" *ngIf="err">⚠️ {{ err }}</div>
        <div class="alert alert-success" *ngIf="ok">{{ ok }}</div>

        <div class="form-group">
          <label class="form-label">{{ 'AUTH.EMAIL' | translate }}</label>
          <input class="form-input" type="email" [(ngModel)]="email" placeholder="you@example.com">
        </div>
        <div class="form-group">
          <label class="form-label">{{ 'AUTH.PASSWORD' | translate }}</label>
          <input class="form-input" type="password" [(ngModel)]="password" placeholder="min 6 chars">
        </div>
        <div class="form-group">
          <label class="form-label">{{ 'AUTH.ROLE' | translate }}</label>
          <div class="role-grid">
            <button class="role-btn" [class.sel]="role==='client'" (click)="role='client'">
              <br><span>Client</span>
            </button>
            <button class="role-btn" [class.sel]="role==='courier'" (click)="role='courier'">
              <br><span>Courier</span>
            </button>
            <button class="role-btn" [class.sel]="role==='dispatcher'" (click)="role='dispatcher'">
              <br><span>Dispatcher</span>
            </button>
            <button class="role-btn" [class.sel]="role==='admin'" (click)="role='admin'">
              <br><span>Admin</span>
            </button>
          </div>
        </div>

        <button class="btn btn-primary w-full btn-lg" (click)="submit()" [disabled]="loading">
          {{ loading ? '⏳' : '' }}
          {{ loading ? ('COMMON.LOADING' | translate) : ('AUTH.REGISTER' | translate) }}
        </button>

        <div class="auth-switch mt-3">
          {{ 'AUTH.HAS_ACC' | translate }}
          <a routerLink="/login" class="text-pink" style="margin-left:4px;font-weight:700">
            {{ 'AUTH.LOGIN' | translate }}
          </a>
        </div>
      </div>
    </div>
  `,
  styles: [`
    .auth-page { min-height:100vh;display:flex;align-items:center;justify-content:center;padding:20px;position:relative;overflow:hidden; }
    .auth-bg { position:fixed;inset:0;pointer-events:none; }
    .float-elem { position:absolute;font-size:64px;opacity:0.07;animation:floatA 7s ease-in-out infinite; }
    .f1{top:10%;left:6%;animation-delay:0s} .f2{top:70%;right:4%;animation-delay:2s} .f3{bottom:15%;left:68%;animation-delay:4s} .f4{top:40%;right:14%;animation-delay:1s}
    @keyframes floatA{0%,100%{transform:translateY(0)}50%{transform:translateY(-20px)}}
    .auth-card { width:100%;max-width:430px;text-align:center;z-index:1;position:relative; }
    .auth-logo { font-size:56px;margin-bottom:6px; }
    .auth-title { font-family:var(--font-h);font-size:26px;font-weight:700;margin-bottom:4px; }
    .auth-sub { color:var(--text2);font-size:14px;margin-bottom:20px; }
    .auth-switch { font-size:14px;color:var(--text2); }
    .role-grid { display:grid;grid-template-columns:repeat(4,1fr);gap:8px; }
    .role-btn {
      background:var(--bg4);border:2px solid var(--border);border-radius:10px;
      padding:12px 4px;cursor:pointer;font-size:20px;text-align:center;transition:all 0.2s;
    }
    .role-btn span { display:block;font-size:11px;font-family:var(--font-h);font-weight:600;margin-top:4px;color:var(--text2); }
    .role-btn.sel { border-color:var(--pink);background:rgba(255,77,166,0.1);color:var(--pink); }
    .role-btn:hover:not(.sel) { background:var(--bg5); }
  `]
})
export class RegisterComponent {
  email = ''; password = ''; role = 'client';
  loading = false; err = ''; ok = '';

  constructor(private auth: AuthService, private router: Router) {}

  submit() {
    if (!this.email || !this.password) { this.err = 'Please fill in all fields'; return; }
    if (this.password.length < 6) { this.err = 'Password must be at least 6 characters'; return; }
    this.loading = true; this.err = ''; this.ok = '';
    this.auth.register({ email: this.email, password: this.password, role: this.role }).subscribe({
      next: () => {
        this.ok = ' Account created! Signing you in...';
        this.auth.login({ email: this.email, password: this.password }).subscribe({
          next: () => this.router.navigate(['/dashboard']),
          error: () => this.router.navigate(['/login'])
        });
      },
      error: e => { this.err = e.error?.error ?? 'Registration failed'; this.loading = false; }
    });
  }
}
