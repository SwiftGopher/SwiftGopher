import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { TranslateModule } from '@ngx-translate/core';
import { forkJoin } from 'rxjs';
import { AuthService } from '../../core/services/auth.service';
import { OrderService } from '../../core/services/order.service';
import { CourierService } from '../../core/services/courier.service';
import { Order } from '../../shared/models/models';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [CommonModule, RouterLink, TranslateModule],
  template: `
    <div class="fade-in">
      <div class="page-header">
        <h1 class="page-title">{{ 'DASH.TITLE' | translate }}</h1>
        <div class="welcome-chip">
          {{ roleEmoji }} <span>{{ auth.currentUser?.email }}</span>
          <span class="badge badge-{{ auth.currentUser?.role }}">{{ auth.currentUser?.role }}</span>
        </div>
      </div>

      <div class="spinner" *ngIf="loading"></div>

      <ng-container *ngIf="!loading">
        <!-- Stats grid for admin/dispatcher -->
        <div class="stats-grid" *ngIf="isAdminOrDispatcher">
          <div class="card stat-card">
            <div class="stat-icon"></div>
            <div class="stat-num">{{ totalOrders }}</div>
            <div class="stat-label">{{ 'DASH.TOTAL' | translate }}</div>
          </div>
          <div class="card stat-card">
            <div class="stat-icon"></div>
            <div class="stat-num" style="color:var(--yellow)">{{ pendingOrders }}</div>
            <div class="stat-label">{{ 'DASH.PENDING' | translate }}</div>
          </div>
          <div class="card stat-card">
            <div class="stat-icon"></div>
            <div class="stat-num" style="color:var(--pink)">{{ inProgressOrders }}</div>
            <div class="stat-label">{{ 'DASH.IN_PROGRESS' | translate }}</div>
          </div>
          <div class="card stat-card">
            <div class="stat-icon"></div>
            <div class="stat-num" style="color:var(--green)">{{ deliveredOrders }}</div>
            <div class="stat-label">{{ 'DASH.DELIVERED' | translate }}</div>
          </div>
          <div class="card stat-card">
            <div class="stat-icon"></div>
            <div class="stat-num" style="color:var(--cyan)">{{ freeCouriers }}</div>
            <div class="stat-label">{{ 'DASH.FREE_C' | translate }}</div>
          </div>
        </div>

        <!-- Stats for courier -->
        <div class="stats-grid grid-courier" *ngIf="isCourier">
          <div class="card stat-card">
            <div class="stat-icon"></div>
            <div class="stat-num" style="color:var(--pink)">{{ myAssigned }}</div>
            <div class="stat-label">{{ 'DASH.PENDING' | translate }}</div>
          </div>
          <div class="card stat-card">
            <div class="stat-icon"></div>
            <div class="stat-num" style="color:var(--green)">{{ myDelivered }}</div>
            <div class="stat-label">{{ 'DASH.DELIVERED' | translate }}</div>
          </div>
        </div>

        <!-- Stats for client -->
        <div class="stats-grid grid-courier" *ngIf="isClient">
          <div class="card stat-card">
            <div class="stat-icon"></div>
            <div class="stat-num">{{ totalOrders }}</div>
            <div class="stat-label">{{ 'DASH.TOTAL' | translate }}</div>
          </div>
          <div class="card stat-card">
            <div class="stat-icon"></div>
            <div class="stat-num" style="color:var(--pink)">{{ inProgressOrders }}</div>
            <div class="stat-label">{{ 'DASH.IN_PROGRESS' | translate }}</div>
          </div>
          <div class="card stat-card">
            <div class="stat-icon"></div>
            <div class="stat-num" style="color:var(--green)">{{ deliveredOrders }}</div>
            <div class="stat-label">{{ 'DASH.DELIVERED' | translate }}</div>
          </div>
        </div>

        <!-- Content grid -->
        <div class="dash-grid">
          <!-- Recent orders -->
          <div class="card" *ngIf="recentOrders.length > 0">
            <div class="flex items-center justify-between mb-3">
              <h2 class="section-title">{{ 'DASH.RECENT' | translate }}</h2>
              <a [routerLink]="isClient ? '/my-orders' : isCourier ? '/my-deliveries' : '/orders'"
                 class="btn btn-sm btn-secondary">
                See all →
              </a>
            </div>
            <div class="recent-list">
              <div class="recent-item" *ngFor="let o of recentOrders">
                <div class="flex items-center justify-between gap-2">
                  <code>{{ o.id.slice(0,8) }}…</code>
                  <span class="badge badge-{{ o.status }}">{{ 'STATUS.' + o.status | translate }}</span>
                </div>
                <div class="route-mini">
                  <span class="text-muted text-sm">{{ o.pickup_address }}</span>
                  <span class="arrow-mini">→</span>
                  <span class="text-muted text-sm">{{ o.delivery_address }}</span>
                </div>
                <div class="flex items-center justify-between">
                  <strong class="text-pink" style="font-family:var(--font-h)">₸{{ o.price }}</strong>
                  <span class="text-sm text-muted">{{ o.created_at | date:'dd.MM HH:mm' }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- Quick actions -->
          <div class="card">
            <h2 class="section-title mb-3">{{ 'DASH.QUICK' | translate }}</h2>
            <div class="qa-grid">
              <a *ngIf="isClient" routerLink="/my-orders" class="qa-btn">
                <span class="qa-icon"></span>
                <span>{{ 'DASH.MY_ORDERS' | translate }}</span>
              </a>
              <a *ngIf="isAdminOrDispatcher" routerLink="/orders" class="qa-btn">
                <span class="qa-icon"></span>
                <span>{{ 'DASH.ALL_ORDERS' | translate }}</span>
              </a>
              <a *ngIf="isAdminOrDispatcher" routerLink="/couriers" class="qa-btn">
                <span class="qa-icon"></span>
                <span>{{ 'DASH.ALL_COURIERS' | translate }}</span>
              </a>
              <a *ngIf="isCourier" routerLink="/my-deliveries" class="qa-btn">
                <span class="qa-icon"></span>
                <span>{{ 'DASH.MY_DELIVERIES' | translate }}</span>
              </a>
              <a *ngIf="isCourier" routerLink="/couriers" class="qa-btn">
                <span class="qa-icon"></span>
                <span>Update Location</span>
              </a>
              <a routerLink="/profile" class="qa-btn">
                <span class="qa-icon"></span>
                <span>{{ 'NAV.PROFILE' | translate }}</span>
              </a>
            </div>
          </div>
        </div>
      </ng-container>
    </div>
  `,
  styles: [`
    .welcome-chip { display:flex;align-items:center;gap:8px;font-size:13px;color:var(--text2);flex-wrap:wrap; }
    .stats-grid { display:grid; grid-template-columns:repeat(auto-fit,minmax(160px,1fr)); gap:14px; margin-bottom:20px; }
    .grid-courier { grid-template-columns:repeat(3,1fr); }
    .dash-grid { display:grid; grid-template-columns:1fr 1fr; gap:18px; }
    .recent-list { display:flex;flex-direction:column;gap:10px; }
    .recent-item { background:var(--bg4); border:1.5px solid var(--border); border-radius:12px; padding:12px; display:flex;flex-direction:column;gap:6px; }
    .route-mini { display:flex;align-items:center;gap:6px;flex-wrap:wrap; }
    .arrow-mini { color:var(--pink);font-weight:700; }
    .qa-grid { display:grid;grid-template-columns:1fr 1fr;gap:10px; }
    .qa-btn {
      display:flex;flex-direction:column;align-items:center;gap:8px;
      background:var(--bg4);border:2px solid var(--border);border-radius:14px;
      padding:18px 10px;text-decoration:none;color:var(--text);
      font-family:var(--font-h);font-size:13px;font-weight:600;text-align:center;
      transition:all 0.2s;cursor:pointer;
    }
    .qa-btn:hover { border-color:var(--pink);color:var(--pink);transform:translateY(-2px);box-shadow:0 4px 12px var(--pink-glow); }
    .qa-icon { font-size:26px; }
    @media(max-width:768px) { .dash-grid{grid-template-columns:1fr} .stats-grid{grid-template-columns:repeat(2,1fr)} .grid-courier{grid-template-columns:repeat(2,1fr)} }
  `]
})
export class DashboardComponent implements OnInit {
  loading = true;
  totalOrders = 0; pendingOrders = 0; inProgressOrders = 0; deliveredOrders = 0;
  freeCouriers = 0; myAssigned = 0; myDelivered = 0;
  recentOrders: Order[] = [];

  constructor(
    public auth: AuthService,
    private orderSvc: OrderService,
    private courierSvc: CourierService
  ) {}

  get isAdminOrDispatcher() { return ['admin','dispatcher'].includes(this.auth.role); }
  get isClient() { return this.auth.role === 'client'; }
  get isCourier() { return this.auth.role === 'courier'; }
  get roleEmoji() { return ({admin:'Admin',dispatcher:'Dispatcher',courier:'Courier',client:'Client'})[this.auth.role]??'Client'; }

  ngOnInit(): void {
    if (this.isAdminOrDispatcher) {
      forkJoin({ orders: this.orderSvc.listOrders({limit:100}), free: this.courierSvc.listFreeCouriers() }).subscribe({
        next: ({orders, free}) => {
          const all = orders.data ?? [];
          this.totalOrders = all.length;
          this.pendingOrders = all.filter(o => o.status === 'pending').length;
          this.inProgressOrders = all.filter(o => o.status === 'in_progress').length;
          this.deliveredOrders = all.filter(o => o.status === 'delivered').length;
          this.freeCouriers = free.length;
          this.recentOrders = all.slice(0, 5);
          this.loading = false;
        },
        error: () => { this.loading = false; }
      });
    } else if (this.isClient) {
      this.orderSvc.getMyOrders().subscribe({
        next: (orders) => {
          const all = orders ?? [];
          this.totalOrders = all.length;
          this.inProgressOrders = all.filter(o => o.status === 'in_progress').length;
          this.deliveredOrders = all.filter(o => o.status === 'delivered').length;
          this.recentOrders = all.slice(0, 5);
          this.loading = false;
        },
        error: () => { this.loading = false; }
      });
    } else if (this.isCourier) {
      // couriers see assigned orders via listOrders (they have access)
      this.orderSvc.listOrders({ limit: 100 }).subscribe({
        next: (res) => {
          const all = res.data ?? [];
          this.myAssigned = all.filter(o => o.status === 'assigned' || o.status === 'in_progress').length;
          this.myDelivered = all.filter(o => o.status === 'delivered').length;
          this.recentOrders = all.slice(0, 5);
          this.loading = false;
        },
        error: () => { this.loading = false; }
      });
    } else {
      this.loading = false;
    }
  }
}
