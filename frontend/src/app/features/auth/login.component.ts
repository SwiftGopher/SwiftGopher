import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { TranslateModule } from '@ngx-translate/core';
import { AuthService } from '../../core/services/auth.service';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterLink, TranslateModule],
  template: `
    <div class="auth-page">
      <div class="auth-bg">
        <div class="float-elem f1">🐹</div>
        <div class="float-elem f2"></div>
        <div class="float-elem f3"></div>
        <div class="float-elem f4"></div>
        <div class="float-elem f5"></div>
      </div>
      <div class="auth-card card slide-up">
        <div class="auth-logo">🐹</div>
        <h1 class="auth-title">{{ 'AUTH.LOGIN_TITLE' | translate }}</h1>
        <p class="auth-sub">{{ 'AUTH.LOGIN_SUB' | translate }}</p>

        <div class="alert alert-error" *ngIf="err">⚠️ {{ err }}</div>

        <div class="form-group">
          <label class="form-label">{{ 'AUTH.EMAIL' | translate }}</label>
          <input class="form-input" type="email" [(ngModel)]="email"
            placeholder="you@swiftgopher.io" (keyup.enter)="submit()">
        </div>
        <div class="form-group">
          <label class="form-label">{{ 'AUTH.PASSWORD' | translate }}</label>
          <input class="form-input" type="password" [(ngModel)]="password"
            placeholder="••••••••" (keyup.enter)="submit()">
        </div>

        <button class="btn btn-primary w-full btn-lg" (click)="submit()" [disabled]="loading">
          {{ loading ? '⏳' : '' }}
          {{ loading ? ('COMMON.LOADING' | translate) : ('AUTH.LOGIN' | translate) }}
        </button>

        <div class="auth-switch mt-3">
          {{ 'AUTH.NO_ACC' | translate }}
          <a routerLink="/register" class="text-pink" style="margin-left:4px;font-weight:700">
            {{ 'AUTH.REGISTER' | translate }}
          </a>
        </div>

        <!-- Demo accounts -->
        <div class="demo-box">
          <div class="demo-label">{{ 'AUTH.DEMO' | translate }}</div>
          <div class="demo-grid">
            <button class="demo-btn" (click)="fill('admin@swiftgopher.io','admin123')">
               Admin<br><span class="demo-hint">admin123</span>
            </button>
            <button class="demo-btn" (click)="fill('dispatcher@swiftgopher.io','disp123')">
               Dispatcher<br><span class="demo-hint">disp123</span>
            </button>
            <button class="demo-btn" (click)="fill('courier1@swiftgopher.io','courier123')">
               Courier<br><span class="demo-hint">courier123</span>
            </button>
            <button class="demo-btn" (click)="fill('client1@swiftgopher.io','client123')">
               Client<br><span class="demo-hint">client123</span>
            </button>
          </div>
          <div class="demo-note">⚠️ Run seed_with_known_passwords.sql in your DB first, or register manually</div>
        </div>
      </div>
    </div>
  `,
  styles: [`
    .auth-page { min-height:100vh;display:flex;align-items:center;justify-content:center;padding:20px;position:relative;overflow:hidden; }
    .auth-bg { position:fixed;inset:0;pointer-events:none; }
    .float-elem { position:absolute;font-size:64px;opacity:0.07;animation:floatAnim 7s ease-in-out infinite; }
    .f1{top:8%;left:4%;animation-delay:0s} .f2{top:65%;left:82%;animation-delay:1.5s} .f3{top:25%;right:6%;animation-delay:3s} .f4{bottom:18%;left:12%;animation-delay:4.5s} .f5{top:50%;left:50%;animation-delay:2s}
    @keyframes floatAnim{0%,100%{transform:translateY(0) rotate(0deg)}50%{transform:translateY(-22px) rotate(8deg)}}
    .auth-card { width:100%;max-width:430px;text-align:center;z-index:1;position:relative; }
    .auth-logo { font-size:64px;margin-bottom:6px;animation:logoBounce 2s ease-in-out infinite; }
    @keyframes logoBounce{0%,100%{transform:translateY(0)}50%{transform:translateY(-8px)}}
    .auth-title { font-family:var(--font-h);font-size:28px;font-weight:700;margin-bottom:4px; }
    .auth-sub { color:var(--text2);font-size:14px;margin-bottom:20px; }
    .auth-switch { font-size:14px;color:var(--text2); }
    .demo-box { border-top:2px dashed var(--border);margin-top:16px;padding-top:14px; }
    .demo-label { font-family:var(--font-h);font-size:11px;font-weight:600;text-transform:uppercase;letter-spacing:0.5px;color:var(--text2);margin-bottom:8px; }
    .demo-grid { display:grid;grid-template-columns:1fr 1fr;gap:6px;margin-bottom:8px; }
    .demo-btn {
      background:var(--bg4);border:2px solid var(--border);border-radius:10px;
      padding:8px 6px;cursor:pointer;font-size:13px;font-family:var(--font-h);
      font-weight:600;color:var(--text2);transition:all 0.2s;line-height:1.4;
    }
    .demo-btn:hover { border-color:var(--pink);color:var(--pink); }
    .demo-hint { font-size:11px;font-weight:400;opacity:0.7;font-family:var(--font-b); }
    .demo-note { font-size:11px;color:var(--text2);opacity:0.7;line-height:1.4; }
  `]
})
export class LoginComponent {
  email = ''; password = ''; loading = false; err = '';

  constructor(private auth: AuthService, private router: Router) {}

  fill(e: string, p: string) { this.email = e; this.password = p; }

  submit() {
    if (!this.email || !this.password) { this.err = 'Please fill in all fields'; return; }
    this.loading = true; this.err = '';
    this.auth.login({ email: this.email, password: this.password }).subscribe({
      next: () => this.router.navigate(['/dashboard']),
      error: e => { this.err = e.error?.error ?? 'Invalid email or password'; this.loading = false; }
    });
  }
}
