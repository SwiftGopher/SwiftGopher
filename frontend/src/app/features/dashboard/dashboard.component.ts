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
          {{ roleEmoji }}
          <span>{{ auth.currentUser?.email }}</span>
          <span class="badge badge-{{ auth.currentUser?.role }}">
            {{ auth.currentUser?.role }}
          </span>
        </div>
      </div>

      <div class="spinner" *ngIf="loading"></div>

      <ng-container *ngIf="!loading">
        
        <div class="stats-grid" *ngIf="isAdminOrDispatcher">
          <div class="card stat-card">
            <div class="stat-num">{{ totalOrders }}</div>
            <div class="stat-label">{{ 'DASH.TOTAL' | translate }}</div>
          </div>

          <div class="card stat-card">
            <div class="stat-num">{{ pendingOrders }}</div>
            <div class="stat-label">{{ 'DASH.PENDING' | translate }}</div>
          </div>

          <div class="card stat-card">
            <div class="stat-num">{{ inProgressOrders }}</div>
            <div class="stat-label">{{ 'DASH.IN_PROGRESS' | translate }}</div>
          </div>

          <div class="card stat-card">
            <div class="stat-num">{{ deliveredOrders }}</div>
            <div class="stat-label">{{ 'DASH.DELIVERED' | translate }}</div>
          </div>

          <div class="card stat-card">
            <div class="stat-num">{{ freeCouriers }}</div>
            <div class="stat-label">{{ 'DASH.FREE_C' | translate }}</div>
          </div>
        </div>
        
        <div class="stats-grid grid-courier" *ngIf="isCourier">
          <div class="card stat-card">
            <div class="stat-num">{{ myAssigned }}</div>
            <div class="stat-label">{{ 'DASH.ASSIGNED' | translate }}</div>
          </div>

          <div class="card stat-card">
            <div class="stat-num">{{ myDelivered }}</div>
            <div class="stat-label">{{ 'DASH.DELIVERED' | translate }}</div>
          </div>
        </div>
        
        <div class="stats-grid grid-courier" *ngIf="isClient">
          <div class="card stat-card">
            <div class="stat-num">{{ totalOrders }}</div>
            <div class="stat-label">{{ 'DASH.TOTAL' | translate }}</div>
          </div>

          <div class="card stat-card">
            <div class="stat-num">{{ inProgressOrders }}</div>
            <div class="stat-label">{{ 'DASH.IN_PROGRESS' | translate }}</div>
          </div>

          <div class="card stat-card">
            <div class="stat-num">{{ deliveredOrders }}</div>
            <div class="stat-label">{{ 'DASH.DELIVERED' | translate }}</div>
          </div>
        </div>
        
        <div class="dash-grid">

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
                <code>{{ o.id.slice(0,8) }}…</code>

                <span class="badge badge-{{ o.status }}">
                  {{ 'STATUS.' + o.status | translate }}
                </span>

                <div class="route-mini">
                  {{ o.pickup_address }} → {{ o.delivery_address }}
                </div>
              </div>
            </div>
          </div>

        </div>

      </ng-container>
    </div>
  `,
  styles: [`
    .stats-grid { display:grid;grid-template-columns:repeat(auto-fit,minmax(160px,1fr));gap:14px; }
    .grid-courier { grid-template-columns:repeat(3,1fr); }
  `]
})
export class DashboardComponent implements OnInit {

  loading = true;

  totalOrders = 0;
  pendingOrders = 0;
  inProgressOrders = 0;
  deliveredOrders = 0;

  freeCouriers = 0;
  myAssigned = 0;
  myDelivered = 0;

  recentOrders: Order[] = [];

  constructor(
      public auth: AuthService,
      private orderSvc: OrderService,
      private courierSvc: CourierService
  ) {}

  get isAdminOrDispatcher() {
    return ['admin', 'dispatcher'].includes(this.auth.role);
  }

  get isClient() {
    return this.auth.role === 'client';
  }

  get isCourier() {
    return this.auth.role === 'courier';
  }

  get roleEmoji() {
    return ({
      admin: 'Admin',
      dispatcher: 'Dispatcher',
      courier: 'Courier',
      client: 'Client'
    } as const)[this.auth.role] ?? 'Client';
  }

  ngOnInit(): void {

    if (this.isAdminOrDispatcher) {
      forkJoin({
        orders: this.orderSvc.listOrders({ limit: 100 }),
        free: this.courierSvc.listFreeCouriers()
      }).subscribe({
        next: ({ orders, free }) => {
          const all = orders.data ?? [];

          this.totalOrders = all.length;
          this.pendingOrders = all.filter(o => o.status === 'pending').length;
          this.inProgressOrders = all.filter(o => o.status === 'in_progress').length;
          this.deliveredOrders = all.filter(o => o.status === 'delivered').length;

          this.freeCouriers = free.length;
          this.recentOrders = all.slice(0, 5);

          this.loading = false;
        },
        error: () => this.loading = false
      });
    }

    else if (this.isClient) {
      this.orderSvc.getMyOrders().subscribe({
        next: (orders) => {
          const all = orders ?? [];

          this.totalOrders = all.length;
          this.inProgressOrders = all.filter(o => o.status === 'in_progress').length;
          this.deliveredOrders = all.filter(o => o.status === 'delivered').length;

          this.recentOrders = all.slice(0, 5);
          this.loading = false;
        },
        error: () => this.loading = false
      });
    }

    else if (this.isCourier) {


      this.orderSvc.getMyCourierOrders().subscribe({
        next: (orders) => {
          const all = orders ?? [];

          this.myAssigned = all.filter(o =>
              o.status === 'assigned' || o.status === 'in_progress'
          ).length;

          this.myDelivered = all.filter(o =>
              o.status === 'delivered'
          ).length;

          this.recentOrders = all.slice(0, 5);

          this.loading = false;
        },
        error: () => this.loading = false
      });
    }

    else {
      this.loading = false;
    }
  }
}
