import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule } from '@ngx-translate/core';
import { AuthService } from '../../core/services/auth.service';
import { ThemeService } from '../../core/services/theme.service';
import { User } from '../../shared/models/models';

@Component({
  selector: 'app-profile',
  standalone: true,
  imports: [CommonModule, TranslateModule],
  template: `
    <div class="fade-in">
      <div class="page-header">
        <h1 class="page-title">{{ 'PROFILE.TITLE' | translate }}</h1>
      </div>

      <div class="profile-layout">
        <!-- Avatar card -->
        <div class="card avatar-card">
          <div class="avatar-ring">
            <span class="avatar-emoji">Panel</span>
          </div>
          <div class="avatar-email">{{ user?.email }}</div>
          <span class="badge badge-{{ user?.role }}" style="font-size:14px;padding:6px 16px">
            {{ user?.role | titlecase }}
          </span>
          <div class="avatar-actions">
            <button class="btn btn-secondary w-full" (click)="theme.toggle()">
              {{ theme.current === 'dark' ? 'Light mode' : 'Dark mode' }}
            </button>
            <button class="btn btn-danger w-full" (click)="auth.logout()">
              {{ 'NAV.LOGOUT' | translate }}
            </button>
          </div>
        </div>

        <!-- Info card -->
        <div class="card">
          <h2 class="section-title mb-4">{{ 'PROFILE.INFO' | translate }}</h2>
          <div class="info-list">
            <div class="info-row">
              <span class="info-lbl">{{ 'PROFILE.EMAIL' | translate }}</span>
              <span class="info-val">{{ user?.email }}</span>
            </div>
            <div class="info-row">
              <span class="info-lbl">{{ 'PROFILE.ROLE' | translate }}</span>
              <span class="badge badge-{{ user?.role }}">{{ user?.role }}</span>
            </div>
            <div class="info-row">
              <span class="info-lbl">User ID</span>
              <code style="font-size:11px;word-break:break-all">{{ user?.id }}</code>
            </div>
            <div class="info-row">
              <span class="info-lbl">{{ 'PROFILE.MEMBER' | translate }}</span>
              <span class="info-val">{{ user?.created_at | date:'dd MMMM yyyy' }}</span>
            </div>
          </div>
        </div>

        <!-- Permissions card -->
        <div class="card">
          <h2 class="section-title mb-4">{{ 'PROFILE.PERMS' | translate }}</h2>
          <div class="perm-list">
            <div class="perm-item" [class.perm-yes]="can('orders:read')">
              <span>{{ can('orders:read') ? 'OK' : 'Bad' }}</span>
              <span>{{ 'PROFILE.PERM_READ' | translate }}</span>
            </div>
            <div class="perm-item" [class.perm-yes]="can('orders:create')">
              <span>{{ can('orders:create') ? 'OK' : 'Bad' }}</span>
              <span>{{ 'PROFILE.PERM_CREATE' | translate }}</span>
            </div>
            <div class="perm-item" [class.perm-yes]="can('orders:update')">
              <span>{{ can('orders:update') ? 'OK' : 'Bad' }}</span>
              <span>{{ 'PROFILE.PERM_UPD' | translate }}</span>
            </div>
            <div class="perm-item" [class.perm-yes]="can('couriers:manage')">
              <span>{{ can('couriers:manage') ? 'OK' : 'Bad' }}</span>
              <span>{{ 'PROFILE.PERM_C_MANAGE' | translate }}</span>
            </div>
            <div class="perm-item" [class.perm-yes]="can('my:orders')">
              <span>{{ can('my:orders') ? 'OK' : 'Bad' }}</span>
              <span>{{ 'PROFILE.PERM_MY' | translate }}</span>
            </div>
            <div class="perm-item" [class.perm-yes]="can('deliver')">
              <span>{{ can('deliver') ? 'OK' : 'Bad' }}</span>
              <span>{{ 'PROFILE.PERM_DELIVER' | translate }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  `,
  styles: [`
    .profile-layout { display:grid;grid-template-columns:240px 1fr 1fr;gap:16px; }
    .avatar-card { display:flex;flex-direction:column;align-items:center;gap:12px;text-align:center; }
    .avatar-ring {
      width:90px;height:90px;border-radius:50%;
      border:3px solid var(--pink);background:var(--bg4);
      display:flex;align-items:center;justify-content:center;
      box-shadow:0 0 20px var(--pink-glow);
    }
    .avatar-emoji { font-size:44px; }
    .avatar-email { font-size:13px;color:var(--text2);word-break:break-all; }
    .avatar-actions { display:flex;flex-direction:column;gap:8px;width:100%; }
    .info-list { display:flex;flex-direction:column; }
    .info-row {
      display:flex;align-items:center;justify-content:space-between;
      padding:12px 0;border-bottom:1.5px solid var(--border);gap:12px;flex-wrap:wrap;
    }
    .info-row:last-child { border-bottom:none; }
    .info-lbl { font-size:13px;font-weight:600;color:var(--text2);white-space:nowrap; }
    .info-val { font-size:13px; }
    .perm-list { display:flex;flex-direction:column;gap:8px; }
    .perm-item {
      display:flex;align-items:center;gap:10px;padding:10px 12px;
      border-radius:10px;background:var(--bg4);border:1.5px solid var(--border);
      font-size:13px;color:var(--text2);
    }
    .perm-item.perm-yes { border-color:var(--green);color:var(--text); }
    @media(max-width:900px) { .profile-layout{grid-template-columns:1fr 1fr} }
    @media(max-width:640px) { .profile-layout{grid-template-columns:1fr} }
  `]
})
export class ProfileComponent implements OnInit {
  user: User | null = null;

  constructor(public auth: AuthService, public theme: ThemeService) {}

  ngOnInit() {
    this.auth.user$.subscribe(u => this.user = u);
    if (!this.user) this.auth.fetchProfile();
  }

  can(perm: string): boolean {
    const r = this.user?.role ?? '';
    const map: Record<string, string[]> = {
      'orders:read':    ['admin','dispatcher','courier'],
      'orders:create':  ['admin','client'],
      'orders:update':  ['admin','dispatcher','courier'],
      'couriers:manage':['admin','dispatcher','courier'],
      'my:orders':      ['client'],
      'deliver':        ['courier'],
    };
    return (map[perm] ?? []).includes(r);
  }
}
