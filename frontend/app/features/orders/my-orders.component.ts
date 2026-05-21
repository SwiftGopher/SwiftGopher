import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { TranslateModule } from '@ngx-translate/core';
import { OrderService } from '../../core/services/order.service';
import { Order, OrderHistory, CreateOrderRequest } from '../../shared/models/models';

@Component({
  selector: 'app-my-orders',
  standalone: true,
  imports: [CommonModule, FormsModule, TranslateModule],
  template: `
    <div class="fade-in">
      <div class="page-header">
        <h1 class="page-title">{{ 'ORDERS.MY' | translate }}</h1>
        <button class="btn btn-primary" (click)="openCreate()">{{ 'ORDERS.CREATE' | translate }}</button>
      </div>

      <div class="spinner" *ngIf="loading"></div>

      <ng-container *ngIf="!loading">
        <div class="empty" *ngIf="orders.length===0">
          <p>{{ 'ORDERS.NONE' | translate }}</p>
          <button class="btn btn-primary mt-3" (click)="openCreate()">{{ 'ORDERS.CREATE' | translate }}</button>
        </div>

        <div class="orders-grid" *ngIf="orders.length>0">
          <div class="order-card card" *ngFor="let o of orders">
            <div class="flex items-center justify-between mb-2">
              <code>{{ o.id.slice(0,8) }}…</code>
              <span class="badge badge-{{ o.status }}">{{ 'STATUS.' + o.status | translate }}</span>
            </div>
            <div class="route-block">
              <div class="route-row">
                <span class="route-dot" style="color:var(--cyan)">●</span>
                <span class="text-sm text-muted">{{ o.pickup_address }}</span>
              </div>
              <div class="route-vline"></div>
              <div class="route-row">
                <span class="route-dot" style="color:var(--pink)">●</span>
                <span class="text-sm text-muted">{{ o.delivery_address }}</span>
              </div>
            </div>
            <div class="flex items-center justify-between mt-2" style="padding-top:10px;border-top:1.5px solid var(--border)">
              <strong class="text-pink" style="font-family:var(--font-h);font-size:17px">₸{{ o.price }}</strong>
              <span class="text-sm text-muted">{{ o.created_at | date:'dd.MM HH:mm' }}</span>
              <button class="btn btn-xs btn-secondary" (click)="openHistory(o)">History</button>
            </div>
          </div>
        </div>
      </ng-container>
    </div>

    <!-- Create modal -->
    <div class="overlay" *ngIf="showCreate" (click)="closeAll()">
      <div class="modal" (click)="$event.stopPropagation()">
        <div class="modal-title">{{ 'ORDERS.CREATE' | translate }}</div>
        <div class="alert alert-success" *ngIf="ok">{{ ok }}</div>
        <div class="alert alert-error" *ngIf="err">{{ err }}</div>
        <div class="form-group">
          <label class="form-label">{{ 'ORDERS.PICKUP_ADDR' | translate }}</label>
          <input class="form-input" [(ngModel)]="req.pickup_address" placeholder="Almaty, Abay 1">
        </div>
        <div class="form-group">
          <label class="form-label">{{ 'ORDERS.DELIVERY_ADDR' | translate }}</label>
          <input class="form-input" [(ngModel)]="req.delivery_address" placeholder="Almaty, Dostyk 10">
        </div>
        <div class="form-group">
          <label class="form-label">{{ 'ORDERS.PRICE_LBL' | translate }}</label>
          <input class="form-input" type="number" [(ngModel)]="req.price" placeholder="500" min="1">
        </div>
        <div class="flex gap-2 mt-3">
          <button class="btn btn-primary flex-1" (click)="submitCreate()" [disabled]="creating">
            {{ 'COMMON.SAVE' | translate }}
          </button>
          <button class="btn btn-secondary" (click)="closeAll()">{{ 'COMMON.CANCEL' | translate }}</button>
        </div>
      </div>
    </div>

    <!-- History modal -->
    <div class="overlay" *ngIf="showHistory && sel" (click)="closeAll()">
      <div class="modal modal-wide" (click)="$event.stopPropagation()">
        <div class="modal-title">Order History</div>
        <div class="text-sm text-muted mb-3">
          {{ sel!.pickup_address }} → {{ sel!.delivery_address }}
        </div>
        <div class="spinner" *ngIf="loadingH"></div>
        <div class="empty" *ngIf="!loadingH && history.length===0" style="padding:20px">
          <p>No history yet</p>
        </div>
        <div class="timeline" *ngIf="history.length>0">
          <div class="tl-item" *ngFor="let h of history; let i=index">
            <div class="tl-dot" style="border-color:var(--pink);color:var(--pink)">{{ i+1 }}</div>
            <div class="tl-content">
              <div class="flex items-center gap-2 flex-wrap">
                <span class="badge badge-{{ h.old_status }}">{{ 'STATUS.' + h.old_status | translate }}</span>
                <span style="color:var(--pink);font-weight:700;font-size:16px">→</span>
                <span class="badge badge-{{ h.new_status }}">{{ 'STATUS.' + h.new_status | translate }}</span>
              </div>
              <div class="text-sm text-muted mt-1">{{ h.changed_at | date:'dd.MM.yyyy HH:mm' }}</div>
            </div>
          </div>
        </div>
        <button class="btn btn-secondary mt-3" (click)="closeAll()">{{ 'COMMON.CLOSE' | translate }}</button>
      </div>
    </div>
  `,
  styles: [`
    .orders-grid { display:grid;grid-template-columns:repeat(auto-fill,minmax(300px,1fr));gap:14px; }
    .order-card { display:flex;flex-direction:column; }
    .route-block { display:flex;flex-direction:column; }
    .route-row { display:flex;align-items:flex-start;gap:8px; }
    .route-dot { font-size:12px;margin-top:3px;flex-shrink:0; }
    .route-vline { width:2px;height:12px;background:var(--border);margin-left:4px; }
  `]
})
export class MyOrdersComponent implements OnInit {
  orders: Order[] = [];
  loading = false;
  showCreate = false; showHistory = false;
  creating = false; ok = ''; err = '';
  req: CreateOrderRequest = { pickup_address:'', delivery_address:'', price:0 };
  sel: Order | null = null;
  history: OrderHistory[] = []; loadingH = false;

  constructor(private svc: OrderService) {}
  ngOnInit() { this.load(); }

  load() {
    this.loading = true;
    this.svc.getMyOrders().subscribe({
      next: (r: Order[]) => { this.orders = r ?? []; this.loading = false; },
      error: () => { this.orders = []; this.loading = false; }
    });
  }

  openCreate() { this.req = {pickup_address:'',delivery_address:'',price:0}; this.ok=''; this.err=''; this.showCreate=true; }

  openHistory(o: Order) {
    this.sel = o; this.history = []; this.loadingH = true; this.showHistory = true; this.showCreate = false;
    this.svc.getHistory(o.id).subscribe({
      next: (h: OrderHistory[]) => { this.history = h ?? []; this.loadingH = false; },
      error: () => { this.loadingH = false; }
    });
  }

  closeAll() { this.showCreate = false; this.showHistory = false; this.sel = null; }

  submitCreate() {
    const { pickup_address, delivery_address, price } = this.req;
    if (!pickup_address || !delivery_address) { this.err = 'Please fill in both addresses'; return; }
    if (!price || price <= 0) { this.err = 'Price must be greater than 0'; return; }
    this.creating = true; this.err = '';
    this.svc.createOrder(this.req).subscribe({
      next: (order: Order) => {
        this.ok = 'Order created!';
        this.orders.unshift(order);
        this.creating = false;
        setTimeout(() => this.closeAll(), 1200);
      },
      error: (e: any) => { this.err = e.error?.error ?? 'Error creating order'; this.creating = false; }
    });
  }
}
