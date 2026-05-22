import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { TranslateModule } from '@ngx-translate/core';
import { OrderService } from '../../core/services/order.service';
import { Order, OrderHistory, OrderStatus } from '../../shared/models/models';

@Component({
  selector: 'app-orders',
  standalone: true,
  imports: [CommonModule, FormsModule, TranslateModule],
  template: `
    <div class="fade-in">
      <div class="page-header">
        <h1 class="page-title">{{ 'ORDERS.TITLE' | translate }}</h1>
        <button class="btn btn-secondary btn-sm" (click)="load()">{{ 'COMMON.REFRESH' | translate }}</button>
      </div>

      <div class="filter-bar">
        <select class="form-select" [(ngModel)]="fs" (change)="load()" style="max-width:180px">
          <option value="">{{ 'ORDERS.ALL' | translate }}</option>
          <option value="pending">Pending</option>
          <option value="assigned">Assigned</option>
          <option value="in_progress">In Progress</option>
          <option value="delivered">Delivered</option>
          <option value="cancelled">Cancelled</option>
        </select>
        <select class="form-select" [(ngModel)]="sortBy" (change)="load()" style="max-width:150px">
          <option value="created_at">By Created</option>
          <option value="price">By Price</option>
          <option value="updated_at">By Updated</option>
        </select>
        <select class="form-select" [(ngModel)]="sortDir" (change)="load()" style="max-width:120px">
          <option value="desc">↓ Newest</option>
          <option value="asc">↑ Oldest</option>
        </select>
        <button class="btn btn-sm btn-secondary" (click)="resetFilters()">{{ 'COMMON.RESET' | translate }}</button>
        <span class="text-sm text-muted" style="margin-left:auto">{{ orders.length }} orders</span>
      </div>

      <div class="spinner" *ngIf="loading"></div>

      <ng-container *ngIf="!loading">
        <div class="empty" *ngIf="orders.length===0">
          <p>{{ 'ORDERS.NONE' | translate }}</p>
        </div>

        <div class="table-wrap" *ngIf="orders.length>0">
          <table>
            <thead>
            <tr>
              <th>{{ 'ORDERS.ID' | translate }}</th>
              <th>{{ 'ORDERS.PICKUP' | translate }}</th>
              <th>{{ 'ORDERS.DELIVERY' | translate }}</th>
              <th>{{ 'ORDERS.STATUS' | translate }}</th>
              <th>Courier</th>
              <th>{{ 'ORDERS.PRICE' | translate }}</th>
              <th>{{ 'ORDERS.CREATED' | translate }}</th>
              <th>{{ 'ORDERS.ACTIONS' | translate }}</th>
            </tr>
            </thead>
            <tbody>
            <tr *ngFor="let o of orders">
              <td><code>{{ o.id.slice(0,8) }}…</code></td>
              <td class="addr-cell" [title]="o.pickup_address">{{ o.pickup_address }}</td>
              <td class="addr-cell" [title]="o.delivery_address">{{ o.delivery_address }}</td>
              <td><span class="badge badge-{{ o.status }}">{{ 'STATUS.' + o.status | translate }}</span></td>
              <td>
                <code *ngIf="o.courier_id" class="courier-chip" [title]="o.courier_id">{{ o.courier_id.slice(0,8) }}…</code>
                <span *ngIf="!o.courier_id" class="text-muted text-sm">—</span>
              </td>
              <td><strong class="text-pink" style="font-family:var(--font-h)">₸{{ o.price }}</strong></td>
              <td class="text-sm text-muted">{{ o.created_at | date:'dd.MM HH:mm' }}</td>
              <td>
                <div class="flex gap-1">
                  <button class="btn btn-xs btn-secondary" (click)="openHistory(o)">History</button>
                  <button class="btn btn-xs btn-primary" (click)="openUpdate(o)">Edit</button>
                </div>
              </td>
            </tr>
            </tbody>
          </table>
        </div>

        <div class="pagination" *ngIf="orders.length>0">
          <button class="btn btn-sm btn-secondary" [disabled]="offset===0" (click)="prev()">← Prev</button>
          <span class="page-num">Page {{ offset/limit+1 }}</span>
          <button class="btn btn-sm btn-secondary" [disabled]="orders.length < limit" (click)="next()">Next →</button>
        </div>
      </ng-container>
    </div>

    <div class="overlay" *ngIf="showUpdate && sel" (click)="closeAll()">
      <div class="modal" (click)="$event.stopPropagation()">
        <div class="modal-title">{{ 'ORDERS.UPDATE' | translate }}</div>
        <div class="text-sm text-muted mb-3">
          <code>{{ sel!.id.slice(0,8) }}…</code> &nbsp;
          Current: <span class="badge badge-{{ sel!.status }}">{{ sel!.status }}</span>
        </div>
        <div class="alert alert-error" *ngIf="err">{{ err }}</div>
        <div class="alert alert-success" *ngIf="ok">{{ ok }}</div>
        <div class="form-group">
          <label class="form-label">{{ 'ORDERS.NEW_STATUS' | translate }}</label>
          <select class="form-select" [(ngModel)]="newStatus">
            <option value="pending">Pending</option>
            <option value="assigned">Assigned</option>
            <option value="in_progress">In Progress</option>
            <option value="delivered">Delivered</option>
            <option value="cancelled">Cancelled</option>
          </select>
        </div>
        <div class="flex gap-2 mt-3">
          <button class="btn btn-primary flex-1" (click)="submitStatus()" [disabled]="saving">
            {{ saving ? 'Saving...' : ('COMMON.SAVE' | translate) }}
          </button>
          <button class="btn btn-secondary" (click)="closeAll()">{{ 'COMMON.CANCEL' | translate }}</button>
        </div>
      </div>
    </div>

    <div class="overlay" *ngIf="showHistory && sel" (click)="closeAll()">
      <div class="modal modal-wide" (click)="$event.stopPropagation()">
        <div class="modal-title">{{ 'ORDERS.HISTORY' | translate }}</div>
        <div class="text-sm text-muted mb-3">
          Order <code>{{ sel!.id.slice(0,8) }}…</code> &nbsp;·&nbsp; {{ sel!.pickup_address }} → {{ sel!.delivery_address }}
        </div>
        <div class="spinner" *ngIf="loadingHistory"></div>
        <div class="empty" *ngIf="!loadingHistory && history.length===0" style="padding:20px">
          <p>{{ 'ORDERS.NO_HISTORY' | translate }}</p>
        </div>
        <div class="timeline" *ngIf="history.length>0">
          <div class="tl-item" *ngFor="let h of history; let i=index">
            <div class="tl-dot" style="border-color:var(--pink);color:var(--pink)">{{ i+1 }}</div>
            <div class="tl-content">
              <div class="flex items-center gap-2 flex-wrap">
                <span class="badge badge-{{ h.old_status }}">{{ 'STATUS.' + h.old_status | translate }}</span>
                <span style="color:var(--pink);font-weight:700;font-size:18px">→</span>
                <span class="badge badge-{{ h.new_status }}">{{ 'STATUS.' + h.new_status | translate }}</span>
              </div>
              <div class="text-sm text-muted mt-1">{{ h.changed_at | date:'dd.MM.yyyy HH:mm:ss' }}</div>
            </div>
          </div>
        </div>
        <button class="btn btn-secondary mt-3" (click)="closeAll()">{{ 'COMMON.CLOSE' | translate }}</button>
      </div>
    </div>
  `,
  styles: [`.addr-cell { max-width:160px; white-space:nowrap; overflow:hidden; text-overflow:ellipsis; font-size:13px; } .courier-chip { font-size:12px; background:var(--bg4); padding:2px 7px; border-radius:6px; color:var(--cyan); letter-spacing:0.3px; cursor:default; }`]
})
export class OrdersComponent implements OnInit {
  orders: Order[] = [];
  loading = false;
  fs: '' | OrderStatus = ''; sortBy = 'created_at'; sortDir: 'desc' | 'asc' = 'desc';
  limit = 20; offset = 0;
  sel: Order | null = null;
  showUpdate = false; showHistory = false;
  newStatus: OrderStatus = 'pending'; saving = false; err = ''; ok = '';
  history: OrderHistory[] = []; loadingHistory = false;

  constructor(private svc: OrderService) {}
  ngOnInit() { this.load(); }

  load() {
    this.loading = true;
    this.svc.listOrders({ status: this.fs || undefined, limit: this.limit, offset: this.offset, sort_by: this.sortBy, sort_dir: this.sortDir }).subscribe({
      next: (r: { data: Order[] }) => { this.orders = r.data ?? []; this.loading = false; },
      error: () => { this.loading = false; }
    });
  }

  resetFilters() { this.fs = ''; this.sortBy = 'created_at'; this.sortDir = 'desc'; this.offset = 0; this.load(); }
  prev() { this.offset = Math.max(0, this.offset - this.limit); this.load(); }
  next() { this.offset += this.limit; this.load(); }

  openUpdate(o: Order) { this.sel = o; this.newStatus = o.status; this.err = ''; this.ok = ''; this.showUpdate = true; this.showHistory = false; }

  openHistory(o: Order) {
    this.sel = o; this.history = []; this.loadingHistory = true;
    this.showHistory = true; this.showUpdate = false;
    this.svc.getHistory(o.id).subscribe({
      next: (h: OrderHistory[]) => { this.history = h ?? []; this.loadingHistory = false; },
      error: () => { this.loadingHistory = false; }
    });
  }

  closeAll() { this.showUpdate = false; this.showHistory = false; this.sel = null; }

  submitStatus() {
    if (!this.sel) return;
    this.saving = true; this.err = '';
    this.svc.updateStatus(this.sel.id, this.newStatus).subscribe({
      next: (updated: Order) => {
        this.ok = 'Status updated!';
        const i = this.orders.findIndex(o => o.id === updated.id);
        if (i >= 0) this.orders[i] = updated;
        this.saving = false;
        setTimeout(() => this.closeAll(), 1200);
      },
      error: (e: any) => { this.err = e.error?.error ?? 'Error'; this.saving = false; }
    });
  }
}
