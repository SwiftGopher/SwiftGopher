import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { TranslateModule } from '@ngx-translate/core';
import { CourierService } from '../../core/services/courier.service';
import { AuthService } from '../../core/services/auth.service';
import { Courier } from '../../shared/models/models';

type Modal = 'status' | 'transport' | 'location' | null;

@Component({
    selector: 'app-couriers',
    standalone: true,
    imports: [CommonModule, FormsModule, TranslateModule],
    template: `
    <div class="fade-in">
      <div class="page-header">
        <h1 class="page-title">{{ 'COURIERS.TITLE' | translate }}</h1>
        <div class="flex gap-2">
          <button class="btn btn-sm" [class.btn-primary]="showFreeOnly" [class.btn-secondary]="!showFreeOnly" (click)="toggleFree()">
            {{ showFreeOnly ? 'All' : 'Free Only' }}
          </button>
          <button class="btn btn-sm btn-secondary" (click)="load()">Refresh</button>
        </div>
      </div>

      <div class="spinner" *ngIf="loading"></div>

      <ng-container *ngIf="!loading">
        <div class="empty" *ngIf="couriers.length===0">
          <p>{{ 'COURIERS.NONE' | translate }}</p>
        </div>

        <div class="couriers-grid" *ngIf="couriers.length>0">
          <div class="courier-card card" *ngFor="let c of couriers">
            <div class="cc-header">
              <div class="cc-icon">{{ transportEmoji(c.transport_type) }}</div>
              <div style="flex:1;min-width:0">
                <div class="flex items-center gap-2 mb-1">
                  <code class="text-sm">{{ c.id.slice(0,8) }}…</code>
                  <span class="badge badge-{{ c.status }}">{{ 'COURIER_STATUS.' + c.status | translate }}</span>
                </div>
                <div class="text-sm text-muted">
                  {{ 'COURIERS.TRANSPORT' | translate }}: <strong>{{ 'COURIERS.TRANSPORT_TYPE.' + c.transport_type | translate }}</strong>
                </div>
              </div>
            </div>

            <div class="cc-location">
              <span *ngIf="c.current_lat !== 0 || c.current_lng !== 0">
                {{ c.current_lat | number:'1.4-4' }}, {{ c.current_lng | number:'1.4-4' }}
              </span>
              <span *ngIf="c.current_lat === 0 && c.current_lng === 0" class="text-muted text-sm">
                Location not set
              </span>
            </div>

            <div class="cc-actions" *ngIf="canEdit">
              <button class="btn btn-xs btn-secondary" (click)="openModal(c,'status')">Status</button>
              <button class="btn btn-xs btn-secondary" (click)="openModal(c,'transport')">Transport</button>
              <button class="btn btn-xs btn-secondary" (click)="openModal(c,'location')">Location</button>
            </div>
          </div>
        </div>
      </ng-container>
    </div>

    <!-- Modal -->
    <div class="overlay" *ngIf="modal && sel" (click)="closeModal()">
      <div class="modal" (click)="$event.stopPropagation()">

        <!-- Status -->
        <ng-container *ngIf="modal==='status'">
          <div class="modal-title">{{ 'COURIERS.UPD_STATUS' | translate }}</div>
          <div class="text-sm text-muted mb-3">Courier <code>{{ sel!.id.slice(0,8) }}…</code></div>
          <div class="alert alert-success" *ngIf="ok">{{ ok }}</div>
          <div class="alert alert-error" *ngIf="err">{{ err }}</div>
          <div class="form-group">
            <label class="form-label">Status</label>
            <select class="form-select" [(ngModel)]="nStatus">
              <option value="free">Free</option>
              <option value="busy">Busy</option>
              <option value="offline">Offline</option>
            </select>
          </div>
          <div class="flex gap-2 mt-3">
            <button class="btn btn-primary flex-1" (click)="saveStatus()" [disabled]="saving">
              {{ saving ? 'Saving...' : ('COMMON.SAVE' | translate) }}
            </button>
            <button class="btn btn-secondary" (click)="closeModal()">{{ 'COMMON.CANCEL' | translate }}</button>
          </div>
        </ng-container>

        <!-- Transport -->
        <ng-container *ngIf="modal==='transport'">
          <div class="modal-title">{{ 'COURIERS.UPD_TRANSPORT' | translate }}</div>
          <div class="text-sm text-muted mb-3">Courier <code>{{ sel!.id.slice(0,8) }}…</code></div>
          <div class="alert alert-success" *ngIf="ok">{{ ok }}</div>
          <div class="alert alert-error" *ngIf="err">{{ err }}</div>
          <div class="transport-grid">
            <button class="transport-opt" [class.selected]="nTransport==='bike'" (click)="nTransport='bike'">
              Bike
            </button>
            <button class="transport-opt" [class.selected]="nTransport==='car'" (click)="nTransport='car'">
              Car
            </button>
            <button class="transport-opt" [class.selected]="nTransport==='foot'" (click)="nTransport='foot'">
              Foot
            </button>
            <button class="transport-opt" [class.selected]="nTransport==='scooter'" (click)="nTransport='scooter'">
              Scooter
            </button>
          </div>
          <div class="flex gap-2 mt-3">
            <button class="btn btn-primary flex-1" (click)="saveTransport()" [disabled]="saving">
              {{ saving ? 'Saving...' : ('COMMON.SAVE' | translate) }}
            </button>
            <button class="btn btn-secondary" (click)="closeModal()">{{ 'COMMON.CANCEL' | translate }}</button>
          </div>
        </ng-container>

        <!-- Location -->
        <ng-container *ngIf="modal==='location'">
          <div class="modal-title">{{ 'COURIERS.UPD_LOCATION' | translate }}</div>
          <div class="text-sm text-muted mb-3">Courier <code>{{ sel!.id.slice(0,8) }}…</code></div>
          <div class="alert alert-success" *ngIf="ok">{{ ok }}</div>
          <div class="alert alert-error" *ngIf="err">{{ err }}</div>
          <div class="grid-2">
            <div class="form-group">
              <label class="form-label">{{ 'COURIERS.LAT' | translate }}</label>
              <input class="form-input" type="number" [(ngModel)]="nLat" step="0.0001" placeholder="43.2220">
            </div>
            <div class="form-group">
              <label class="form-label">{{ 'COURIERS.LNG' | translate }}</label>
              <input class="form-input" type="number" [(ngModel)]="nLng" step="0.0001" placeholder="76.8512">
            </div>
          </div>
          <div class="mb-3">
            <div class="form-label mb-2">{{ 'COURIERS.PRESETS' | translate }}</div>
            <div class="flex gap-2 flex-wrap">
              <button class="btn btn-xs btn-secondary" (click)="setPreset(51.1801,71.4460)">Astana</button>
              <button class="btn btn-xs btn-secondary" (click)="setPreset(43.2220,76.8512)">Almaty</button>
              <button class="btn btn-xs btn-secondary" (click)="setPreset(42.3400,69.5900)">Shymkent</button>
              <button class="btn btn-xs btn-secondary" (click)="setPreset(51.5074,-0.1278)">London</button>
            </div>
          </div>
          <div class="flex gap-2">
            <button class="btn btn-primary flex-1" (click)="saveLocation()" [disabled]="saving">
              {{ saving ? 'Saving...' : ('COMMON.SAVE' | translate) }}
            </button>
            <button class="btn btn-secondary" (click)="closeModal()">{{ 'COMMON.CANCEL' | translate }}</button>
          </div>
        </ng-container>

      </div>
    </div>
  `,
    styles: [`
    .couriers-grid { display:grid;grid-template-columns:repeat(auto-fill,minmax(280px,1fr));gap:14px; }
    .courier-card { display:flex;flex-direction:column;gap:10px; }
    .cc-header { display:flex;align-items:flex-start;gap:12px; }
    .cc-icon { font-size:32px;background:var(--bg4);border:2px solid var(--border);border-radius:10px;padding:8px;flex-shrink:0;text-align:center;min-width:52px; }
    .cc-location { font-size:13px;color:var(--text2); }
    .cc-actions { display:flex;gap:6px;flex-wrap:wrap;padding-top:8px;border-top:1.5px solid var(--border); }
    .transport-grid { display:grid;grid-template-columns:repeat(4,1fr);gap:8px;margin-bottom:4px; }
    .transport-opt {
      background:var(--bg4);border:2px solid var(--border);border-radius:10px;
      padding:12px 6px;cursor:pointer;font-size:20px;text-align:center;
      transition:all 0.2s;color:var(--text2);font-family:var(--font-h);font-size:13px;
    }
    .transport-opt span { display:block;font-size:11px;margin-top:4px; }
    .transport-opt.selected { border-color:var(--pink);background:rgba(255,77,166,0.1);color:var(--pink); }
    .transport-opt:hover:not(.selected) { border-color:var(--border);background:var(--bg5); }
  `]
})
export class CouriersComponent implements OnInit {
    couriers: Courier[] = [];
    loading = false;
    showFreeOnly = false;
    modal: Modal = null;
    sel: Courier | null = null;
    nStatus = 'free'; nTransport = 'bike'; nLat = 0; nLng = 0;
    saving = false; ok = ''; err = '';

    constructor(private svc: CourierService, public auth: AuthService) {}
    get canEdit() { return ['admin','dispatcher','courier'].includes(this.auth.role); }
    ngOnInit() { this.load(); }

    load() {
        this.loading = true;
        const req = this.showFreeOnly ? this.svc.listFreeCouriers() : this.svc.listCouriers();
        req.subscribe({ next: c => { this.couriers = c ?? []; this.loading = false; }, error: () => { this.loading = false; } });
    }

    toggleFree() { this.showFreeOnly = !this.showFreeOnly; this.load(); }
    transportEmoji(t: string) { return ({bike:'🚲',car:'🚗',foot:'🚶',scooter:'🛵'})[t] ?? '🚚'; }

    openModal(c: Courier, m: Modal) {
        this.sel = c; this.modal = m; this.ok = ''; this.err = '';
        if (m === 'status') this.nStatus = c.status;
        if (m === 'transport') this.nTransport = c.transport_type;
        if (m === 'location') { this.nLat = c.current_lat; this.nLng = c.current_lng; }
    }

    closeModal() { this.modal = null; this.sel = null; }
    setPreset(lat: number, lng: number) { this.nLat = lat; this.nLng = lng; }

    saveStatus() {
        if (!this.sel) return; this.saving = true; this.err = '';
        this.svc.updateStatus(this.sel.id, this.nStatus).subscribe({
            next: c => { this.ok = 'Status updated!'; this.patch(c); this.saving = false; setTimeout(() => this.closeModal(), 1000); },
            error: e => { this.err = e.error?.error ?? 'Error'; this.saving = false; }
        });
    }

    saveTransport() {
        if (!this.sel) return; this.saving = true; this.err = '';
        this.svc.updateTransport(this.sel.id, this.nTransport).subscribe({
            next: c => { this.ok = 'Transport updated!'; this.patch(c); this.saving = false; setTimeout(() => this.closeModal(), 1000); },
            error: e => { this.err = e.error?.error ?? 'Error'; this.saving = false; }
        });
    }

    saveLocation() {
        if (!this.sel) return; this.saving = true; this.err = '';
        this.svc.updateLocation(this.sel.id, this.nLat, this.nLng).subscribe({
            next: c => { this.ok = 'Location updated!'; this.patch(c); this.saving = false; setTimeout(() => this.closeModal(), 1000); },
            error: e => { this.err = e.error?.error ?? 'Error'; this.saving = false; }
        });
    }

    private patch(updated: Courier) { const i = this.couriers.findIndex(c => c.id === updated.id); if (i >= 0) this.couriers[i] = updated; }
}