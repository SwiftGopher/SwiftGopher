import { Component, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterOutlet, RouterLink, RouterLinkActive, Router } from '@angular/router';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { Subscription } from 'rxjs';
import { AuthService } from './core/services/auth.service';
import { ThemeService } from './core/services/theme.service';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [CommonModule, RouterOutlet, RouterLink, RouterLinkActive, TranslateModule],
  template: `
    <div [attr.data-theme]="theme.current">

      <!-- ── Desktop Sidebar ── -->
      <aside class="sidebar" *ngIf="auth.isLoggedIn">
        <div class="sb-top">
          <div class="sb-brand">
            <span class="sb-gopher">🐹</span>
            <span class="sb-name">SwiftGopher</span>
          </div>

          <nav class="sb-nav">
            <a routerLink="/dashboard" routerLinkActive="active" class="sb-link">
              <span class="sb-icon">🏠</span>
              <span class="sb-txt">{{ 'NAV.DASHBOARD' | translate }}</span>
            </a>

            <!-- Admin / Dispatcher -->
            <ng-container *ngIf="isAdminOrDispatcher">
              <a routerLink="/orders" routerLinkActive="active" class="sb-link">
                <span class="sb-icon">📦</span>
                <span class="sb-txt">{{ 'NAV.ORDERS' | translate }}</span>
              </a>
              <a routerLink="/couriers" routerLinkActive="active" class="sb-link">
                <span class="sb-icon">🛵</span>
                <span class="sb-txt">{{ 'NAV.COURIERS' | translate }}</span>
              </a>
            </ng-container>

            <!-- Client -->
            <ng-container *ngIf="isClient">
              <a routerLink="/my-orders" routerLinkActive="active" class="sb-link">
                <span class="sb-icon">📦</span>
                <span class="sb-txt">{{ 'NAV.MY_ORDERS' | translate }}</span>
              </a>
            </ng-container>

            <!-- Courier -->
            <ng-container *ngIf="isCourier">
              <a routerLink="/my-deliveries" routerLinkActive="active" class="sb-link">
                <span class="sb-icon">🚚</span>
                <span class="sb-txt">{{ 'NAV.MY_DELIVERIES' | translate }}</span>
              </a>
              <a routerLink="/couriers" routerLinkActive="active" class="sb-link">
                <span class="sb-icon">📍</span>
                <span class="sb-txt">{{ 'NAV.COURIERS' | translate }}</span>
              </a>
            </ng-container>

            <a routerLink="/profile" routerLinkActive="active" class="sb-link">
              <span class="sb-icon">👤</span>
              <span class="sb-txt">{{ 'NAV.PROFILE' | translate }}</span>
            </a>
          </nav>
        </div>

        <div class="sb-bottom">
          <!-- Language -->
          <div class="lang-row">
            <button class="lang-btn" [class.active]="lang==='en'" (click)="setLang('en')">EN</button>
            <button class="lang-btn" [class.active]="lang==='ru'" (click)="setLang('ru')">RU</button>
            <button class="lang-btn" [class.active]="lang==='kz'" (click)="setLang('kz')">KZ</button>
          </div>
          <!-- Theme -->
          <button class="theme-toggle" (click)="theme.toggle()">
            {{ theme.current === 'dark' ? '☀️ Light' : '🌙 Dark' }}
          </button>
          <!-- User -->
          <div class="sb-user">
            <div class="sb-avatar">{{ roleEmoji }}</div>
            <div class="sb-user-info">
              <div class="sb-email">{{ auth.currentUser?.email }}</div>
              <span class="badge badge-{{ auth.currentUser?.role }}">{{ auth.currentUser?.role }}</span>
            </div>
          </div>
          <button class="btn btn-danger btn-sm w-full" (click)="auth.logout()">
            🚪 {{ 'NAV.LOGOUT' | translate }}
          </button>
        </div>
      </aside>

      <!-- ── Main ── -->
      <main class="main" [class.with-sidebar]="auth.isLoggedIn">

        <!-- Mobile topbar -->
        <div class="mobile-top" *ngIf="auth.isLoggedIn">
          <div class="mob-brand">
            <span>🐹</span>
            <span style="font-family:var(--font-h);font-weight:700;font-size:18px;background:linear-gradient(135deg,var(--text),var(--pink));-webkit-background-clip:text;-webkit-text-fill-color:transparent;background-clip:text">SwiftGopher</span>
          </div>
          <div style="display:flex;gap:8px;align-items:center">
            <button class="lang-btn" [class.active]="lang==='en'" (click)="setLang('en')">EN</button>
            <button class="lang-btn" [class.active]="lang==='ru'" (click)="setLang('ru')">RU</button>
            <button class="lang-btn" [class.active]="lang==='kz'" (click)="setLang('kz')">KZ</button>
            <button class="btn btn-sm btn-secondary" (click)="theme.toggle()">
              {{ theme.current === 'dark' ? '☀️' : '🌙' }}
            </button>
          </div>
        </div>

        <div class="page-wrap">
          <router-outlet></router-outlet>
        </div>
      </main>

      <!-- Mobile bottom nav -->
      <nav class="mob-nav" *ngIf="auth.isLoggedIn">
        <a routerLink="/dashboard" routerLinkActive="active" class="mob-item">🏠</a>
        <ng-container *ngIf="isAdminOrDispatcher">
          <a routerLink="/orders" routerLinkActive="active" class="mob-item">📦</a>
          <a routerLink="/couriers" routerLinkActive="active" class="mob-item">🛵</a>
        </ng-container>
        <a *ngIf="isClient" routerLink="/my-orders" routerLinkActive="active" class="mob-item">📦</a>
        <ng-container *ngIf="isCourier">
          <a routerLink="/my-deliveries" routerLinkActive="active" class="mob-item">🚚</a>
          <a routerLink="/couriers" routerLinkActive="active" class="mob-item">📍</a>
        </ng-container>
        <a routerLink="/profile" routerLinkActive="active" class="mob-item">👤</a>
      </nav>

    </div>
  `,
  styles: [`
    :host { display: block; }

    /* ── Sidebar ── */
    .sidebar {
      position: fixed; left: 0; top: 0; bottom: 0; width: 230px;
      background: var(--bg2); border-right: 2.5px solid var(--border);
      display: flex; flex-direction: column; z-index: 50;
      overflow-y: auto;
    }
    .sb-top { flex: 1; padding: 16px 12px; }
    .sb-brand {
      display: flex; align-items: center; gap: 8px;
      padding: 8px 8px 20px;
    }
    .sb-gopher { font-size: 28px; animation: sbBounce 2.5s ease-in-out infinite; }
    @keyframes sbBounce { 0%,100%{transform:translateY(0)} 50%{transform:translateY(-4px)} }
    .sb-name {
      font-family: var(--font-h); font-size: 20px; font-weight: 700;
      background: linear-gradient(135deg, var(--text), var(--pink));
      -webkit-background-clip: text; -webkit-text-fill-color: transparent; background-clip: text;
    }
    .sb-nav { display: flex; flex-direction: column; gap: 3px; }
    .sb-link {
      display: flex; align-items: center; gap: 10px;
      padding: 9px 12px; border-radius: 10px; text-decoration: none;
      color: var(--text2); font-family: var(--font-h); font-size: 15px; font-weight: 600;
      transition: all 0.15s; border: 2px solid transparent;
    }
    .sb-link:hover { background: var(--bg4); color: var(--text); border-color: var(--border); }
    .sb-link.active { background: var(--bg4); color: var(--pink); border-color: var(--pink); }
    .sb-icon { font-size: 18px; width: 22px; text-align: center; }
    .sb-bottom { padding: 12px; border-top: 2px solid var(--border); display: flex; flex-direction: column; gap: 8px; }
    .lang-row { display: flex; gap: 4px; background: var(--bg4); padding: 4px; border-radius: 10px; }
    .lang-btn {
      flex: 1; font-family: var(--font-h); font-size: 12px; font-weight: 700;
      background: none; border: none; cursor: pointer; padding: 4px;
      border-radius: 7px; color: var(--text2); transition: all 0.15s;
    }
    .lang-btn.active { background: var(--pink); color: #fff; }
    .theme-toggle {
      width: 100%; font-family: var(--font-h); font-size: 13px; font-weight: 600;
      background: var(--bg4); border: 2px solid var(--border); border-radius: 9px;
      padding: 7px; cursor: pointer; color: var(--text2); transition: all 0.15s;
    }
    .theme-toggle:hover { border-color: var(--pink); color: var(--pink); }
    .sb-user { display: flex; align-items: center; gap: 8px; padding: 8px; background: var(--bg4); border-radius: 10px; border: 2px solid var(--border); }
    .sb-avatar { font-size: 22px; width: 32px; height: 32px; display:flex;align-items:center;justify-content:center; background: var(--bg5); border-radius: 8px; }
    .sb-user-info { flex: 1; min-width: 0; }
    .sb-email { font-size: 11px; color: var(--text2); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; margin-bottom: 2px; }

    /* ── Main ── */
    .main { min-height: 100vh; }
    .main.with-sidebar { margin-left: 230px; }
    .page-wrap { padding: 20px; max-width: 1100px; }

    /* ── Mobile top ── */
    .mobile-top {
      display: none; position: sticky; top: 0; z-index: 40;
      background: var(--bg2); border-bottom: 2px solid var(--border);
      padding: 10px 16px; align-items: center; justify-content: space-between;
    }
    .mob-brand { display: flex; align-items: center; gap: 6px; }

    /* ── Mobile bottom nav ── */
    .mob-nav {
      display: none; position: fixed; bottom: 0; left: 0; right: 0;
      background: var(--bg2); border-top: 2.5px solid var(--border);
      justify-content: space-around; padding: 8px 0; z-index: 50;
    }
    .mob-item {
      font-size: 22px; padding: 6px 12px; border-radius: 10px;
      text-decoration: none; transition: transform 0.2s;
      border: 2px solid transparent;
    }
    .mob-item.active { border-color: var(--pink); transform: scale(1.1); }

    @media (max-width: 768px) {
      .sidebar { display: none; }
      .main.with-sidebar { margin-left: 0; }
      .mobile-top { display: flex; }
      .mob-nav { display: flex; }
      .page-wrap { padding: 12px 12px 80px; }
    }
  `]
})
export class AppComponent implements OnInit, OnDestroy {
  lang = 'en';
  private sub?: Subscription;

  constructor(
    public auth: AuthService,
    public theme: ThemeService,
    private translate: TranslateService
  ) {}

  ngOnInit(): void {
    const saved = localStorage.getItem('sg_lang') ?? 'en';
    this.translate.setDefaultLang('en');
    this.translate.use(saved);
    this.lang = saved;
  }

  ngOnDestroy(): void { this.sub?.unsubscribe(); }

  setLang(l: string): void {
    this.translate.use(l);
    localStorage.setItem('sg_lang', l);
    this.lang = l;
  }

  get isAdminOrDispatcher(): boolean { return ['admin','dispatcher'].includes(this.auth.role); }
  get isClient(): boolean { return this.auth.role === 'client'; }
  get isCourier(): boolean { return this.auth.role === 'courier'; }

  get roleEmoji(): string {
    return ({admin:'👑',dispatcher:'📋',courier:'🛵',client:'👤'})[this.auth.role] ?? '👤';
  }
}
