import { Component, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { TranslateModule } from '@ngx-translate/core';
import { Subject, Subscription, of } from 'rxjs';
import { debounceTime, distinctUntilChanged, switchMap, catchError } from 'rxjs/operators';

import { OrderService } from '../../core/services/order.service';
import { GeocodingService } from '../../core/services/geocoding.service';
import { Order, OrderHistory, CreateOrderRequest } from '../../shared/models/models';

@Component({
  selector: 'app-my-orders',
  standalone: true,
  imports: [CommonModule, FormsModule, TranslateModule],

  template: `
  <div class="fade-in">

  
    <div class="page-header">
      <h1 class="page-title">{{ 'ORDERS.MY' | translate }}</h1>
      <button class="btn btn-primary" (click)="openCreate()">
        + {{ 'ORDERS.CREATE' | translate }}
      </button>
    </div>

    <!-- Loading -->
    <div class="spinner" *ngIf="loading"></div>

    <ng-container *ngIf="!loading">

      <!-- Empty -->
      <div class="empty" *ngIf="orders.length === 0">
        <span class="empty-emoji"></span>
        <p>{{ 'ORDERS.NONE' | translate }}</p>
      </div>

      <!-- Orders Grid -->
      <div class="orders-grid" *ngIf="orders.length > 0">
        <div class="card order-card" *ngFor="let o of orders">

          <div class="flex items-center justify-between mb-2">
            <code class="order-id">{{ o.id?.slice(0,8) }}…</code>
            <span class="badge badge-{{ o.status }}">{{ o.status }}</span>
          </div>

          <div class="route-block">
            <div class="route-row">
              <span class="route-dot pickup"></span>
              <span class="route-text">{{ o.pickup_address || '—' }}</span>
            </div>
            <div class="route-line"></div>
            <div class="route-row">
              <span class="route-dot delivery"></span>
              <span class="route-text">{{ o.delivery_address || '—' }}</span>
            </div>
          </div>

          <div class="flex items-center justify-between mt-2">
            <strong class="price-tag">₸{{ o.price ?? 0 }}</strong>
            <button class="btn btn-xs btn-secondary" (click)="openHistory(o)">
              History
            </button>
          </div>

        </div>
      </div>
    </ng-container>

    <!-- Create Modal -->
    <div class="overlay" *ngIf="showCreate" (click)="closeAll()">
      <div class="modal" (click)="$event.stopPropagation()">

        <div class="flex items-center justify-between mb-2">
          <h2 class="modal-title">Create Order</h2>
          <button class="modal-close" (click)="closeAll()"></button>
        </div>

        <div class="alert alert-error" *ngIf="err"> {{ err }}</div>
        <div class="alert alert-success" *ngIf="ok"> {{ ok }}</div>

        <!-- Pickup Address -->
        <div class="form-group">
          <label class="form-label">Pickup Address</label>
          <input
            class="form-input"
            [(ngModel)]="req.pickup_address"
            (ngModelChange)="onPickupAddressChange($event)"
            placeholder="Almaty, Abay 1"
          />
          <div class="geo-status" *ngIf="geocoding">
            <span class="spinner-sm spinner"></span> Searching…
          </div>
          <div class="geo-status success" *ngIf="geocoded">
             Found — {{ req.pickup_lat | number:'1.5-5' }}, {{ req.pickup_lng | number:'1.5-5' }}
            <span class="edit-coords" (click)="showCoords = true">edit</span>
          </div>
          <div class="geo-status error" *ngIf="geoFailed"> Not found — enter coordinates manually</div>
        </div>

        <!-- Manual coords -->
        <div class="form-row" *ngIf="geoFailed || showCoords">
          <div class="form-group" style="flex:1">
            <label class="form-label">Latitude</label>
            <input type="number" class="form-input" placeholder="43.2565" [(ngModel)]="req.pickup_lat">
          </div>
          <div class="form-group" style="flex:1">
            <label class="form-label">Longitude</label>
            <input type="number" class="form-input" placeholder="76.9286" [(ngModel)]="req.pickup_lng">
          </div>
        </div>

        <!-- Delivery Address -->
        <div class="form-group">
          <label class="form-label">Delivery Address</label>
          <input class="form-input" [(ngModel)]="req.delivery_address" placeholder="Almaty, Dostyk 5">
        </div>

        <!-- Price -->
        <div class="form-group">
          <label class="form-label">Price (₸)</label>
          <input type="number" class="form-input" [(ngModel)]="req.price" placeholder="500">
        </div>

        <div class="flex gap-2">
          <button class="btn btn-primary"
                  (click)="submitCreate()"
                  [disabled]="creating || geocoding">
            {{ creating ? 'Saving…' : 'Save Order' }}
          </button>
          <button class="btn btn-secondary" (click)="closeAll()">Cancel</button>
        </div>

      </div>
    </div>

    <!-- History Modal -->
    <div class="overlay" *ngIf="showHistory" (click)="closeAll()">
      <div class="modal" (click)="$event.stopPropagation()">

        <div class="flex items-center justify-between mb-2">
          <h2 class="modal-title">Order History</h2>
          <button class="modal-close" (click)="closeAll()"></button>
        </div>

        <div class="spinner" *ngIf="loadingH"></div>

        <div class="empty" *ngIf="history.length === 0 && !loadingH">
          <p>No history yet</p>
        </div>

        <div class="history-list">
          <div class="history-row" *ngFor="let h of history; let i = index">
            <span class="history-num">{{ i + 1 }}</span>
            <span class="badge badge-{{ h.old_status }}">{{ h.old_status }}</span>
            <span class="history-arrow">→</span>
            <span class="badge badge-{{ h.new_status }}">{{ h.new_status }}</span>
            <span class="text-sm text-muted" style="margin-left:auto">
              {{ h.changed_at | date:'dd.MM.yy HH:mm' }}
            </span>
          </div>
        </div>

        <div class="flex" style="margin-top:16px">
          <button class="btn btn-secondary" (click)="closeAll()">Close</button>
        </div>

      </div>
    </div>

  </div>
  `,

  styles: [`
    .orders-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
      gap: 16px;
    }
    .order-card { padding: 16px; }
    .order-id {
      font-size: 12px;
      background: var(--bg4);
      padding: 2px 8px;
      border-radius: 6px;
      color: var(--cyan);
      letter-spacing: 0.5px;
    }

    /* Route */
    .route-block {
      margin: 12px 0;
      padding: 10px 12px;
      background: var(--bg4);
      border-radius: 10px;
      border: 1px solid var(--border);
    }
    .route-row { display: flex; align-items: flex-start; gap: 8px; }
    .route-dot {
      width: 10px; height: 10px; border-radius: 50%;
      margin-top: 4px; flex-shrink: 0;
    }
    .route-dot.pickup   { background: var(--cyan); box-shadow: 0 0 6px var(--cyan); }
    .route-dot.delivery { background: var(--pink); box-shadow: 0 0 6px var(--pink); }
    .route-line {
      width: 2px; height: 14px;
      background: var(--border);
      margin: 3px 0 3px 4px;
    }
    .route-text { font-size: 13px; color: var(--text); line-height: 1.4; }

    .price-tag {
      font-family: var(--font-h);
      font-size: 17px;
      color: var(--green);
    }

    /* Geo status */
    .geo-status {
      font-size: 12px; margin-top: 4px;
      display: flex; align-items: center; gap: 6px;
      color: var(--text2);
    }
    .geo-status.success { color: var(--green); }
    .geo-status.error   { color: var(--red); }
    .edit-coords {
      cursor: pointer; color: var(--cyan);
      text-decoration: underline; margin-left: 6px; font-size: 11px;
    }
    .form-row { display: flex; gap: 10px; }

    /* History */
    .history-list { display: flex; flex-direction: column; gap: 8px; }
    .history-row {
      display: flex; align-items: center; gap: 8px;
      padding: 8px 12px;
      background: var(--bg4);
      border-radius: 8px;
      border: 1px solid var(--border);
    }
    .history-num {
      font-size: 11px; color: var(--text2);
      min-width: 18px; text-align: center;
    }
    .history-arrow { color: var(--text2); font-size: 14px; }
  `]
})
export class MyOrdersComponent implements OnInit, OnDestroy {

  orders: Order[] = [];
  loading = false;

  showCreate = false;
  showHistory = false;

  creating = false;
  ok = '';
  err = '';

  sel: Order | null = null;
  history: OrderHistory[] = [];
  loadingH = false;

  showCoords = false;

  req: CreateOrderRequest = {
    pickup_address: '',
    delivery_address: '',
    pickup_lat: null as any,
    pickup_lng: null as any,
    price: 0
  };

  geocoding = false;
  geocoded = false;
  geoFailed = false;

  private addressInput$ = new Subject<string>();
  private sub?: Subscription;

  constructor(
      private svc: OrderService,
      private geo: GeocodingService
  ) {}

  ngOnInit() {
    this.load();

    this.sub = this.addressInput$.pipe(
        debounceTime(500),
        distinctUntilChanged(),
        switchMap(addr => {
          this.geocoding = true;
          this.geocoded = false;
          this.geoFailed = false;
          return this.geo.geocode(addr).pipe(catchError(() => of(null)));
        })
    ).subscribe(res => {
      this.geocoding = false;
      if (res?.lat && res?.lng) {
        this.req.pickup_lat = res.lat;
        this.req.pickup_lng = res.lng;
        this.geocoded = true;
      } else {
        this.geoFailed = true;
        this.req.pickup_lat = null as any;
        this.req.pickup_lng = null as any;
      }
    });
  }

  ngOnDestroy() { this.sub?.unsubscribe(); }

  onPickupAddressChange(val: string) {
    if (val && val.trim().length >= 3) this.addressInput$.next(val.trim());
  }

  load() {
    this.loading = true;
    this.svc.getMyOrders().subscribe({
      next: r => { this.orders = r ?? []; this.loading = false; },
      error: () => { this.loading = false; }
    });
  }

  openCreate() {
    this.req = { pickup_address: '', delivery_address: '', pickup_lat: null as any, pickup_lng: null as any, price: 0 };
    this.showCreate = true;
    this.err = ''; this.ok = '';
    this.geocoded = false; this.geoFailed = false; this.showCoords = false;
  }

  closeAll() { this.showCreate = false; this.showHistory = false; }

  openHistory(o: Order) {
    this.sel = o;
    this.history = [];
    this.showHistory = true;
    this.loadingH = true;
    this.svc.getHistory(o.id).subscribe({
      next: h => { this.history = h ?? []; this.loadingH = false; },
      error: () => { this.history = []; this.loadingH = false; }
    });
  }

  submitCreate() {
    if (!this.req.pickup_address || !this.req.delivery_address) { this.err = 'Fill in both addresses'; return; }
    if (this.req.pickup_lat == null || this.req.pickup_lng == null) { this.err = 'Pickup coordinates are required'; return; }
    if (this.req.price <= 0) { this.err = 'Price must be greater than 0'; return; }

    this.creating = true;
    this.svc.createOrder(this.req).subscribe({
      next: o => { this.orders.unshift(o); this.creating = false; this.showCreate = false; },
      error: e => { this.err = e?.error?.error ?? 'Something went wrong'; this.creating = false; }
    });
  }
}
