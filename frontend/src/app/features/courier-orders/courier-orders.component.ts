import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { TranslateModule } from '@ngx-translate/core';
import { OrderService } from '../../core/services/order.service';
import { CourierService } from '../../core/services/courier.service';
import { AuthService } from '../../core/services/auth.service';
import { Order, OrderHistory, Courier } from '../../shared/models/models';

@Component({
    selector: 'app-courier-orders',
    standalone: true,
    imports: [CommonModule, FormsModule, TranslateModule],
    template: `
        <div class="fade-in">
            <div class="page-header">
                <h1 class="page-title"> {{ 'COURIERS.MY_ORDERS' | translate }}</h1>
                <div class="flex gap-2 items-center">
          <span *ngIf="myCourier" class="badge badge-{{ myCourier.status }}">
            {{ 'COURIER_STATUS.' + myCourier.status | translate }}
          </span>
                    <button class="btn btn-sm btn-secondary" (click)="loadAll()"></button>
                </div>
            </div>

            <!-- My courier info card -->
            <div class="courier-info-card card mb-4" *ngIf="myCourier">
                <div class="cinfo-grid">
                    <div class="cinfo-item">
                        <div class="cinfo-label">My Status</div>
                        <span class="badge badge-{{ myCourier.status }}" style="font-size:14px;padding:5px 12px">
              {{ 'COURIER_STATUS.' + myCourier.status | translate }}
            </span>
                    </div>
                    <div class="cinfo-item">
                        <div class="cinfo-label">Transport</div>
                        <span class="text-pink font-h fw-7">{{ transportEmoji(myCourier.transport_type) }} {{ myCourier.transport_type }}</span>
                    </div>
                    <div class="cinfo-item">
                        <div class="cinfo-label">Location</div>
                        <span class="text-sm text-muted">
              {{ myCourier.current_lat !== 0 ? (myCourier.current_lat | number:'1.2-4') + ', ' + (myCourier.current_lng | number:'1.2-4') : 'Not set' }}
            </span>
                    </div>
                </div>
                <div class="flex gap-2 flex-wrap mt-3">
                    <button class="btn btn-sm btn-secondary" (click)="openMyStatus()"> Change Status</button>
                    <button class="btn btn-sm btn-secondary" (click)="openMyTransport()"> Change Transport</button>
                    <button class="btn btn-sm btn-secondary" (click)="openMyLocation()"> Update Location</button>
                </div>
            </div>

            <!-- Tabs -->
            <div class="tabs">
                <button class="tab-btn" [class.active]="tab==='active'" (click)="tab='active'">
                     Active ({{ activeOrders.length }})
                </button>
                <button class="tab-btn" [class.active]="tab==='history'" (click)="tab='history'">
                     Completed ({{ completedOrders.length }})
                </button>
                <button class="tab-btn" [class.active]="tab==='all'" (click)="tab='all'">
                     All ({{ allOrders.length }})
                </button>
            </div>

            <div class="spinner" *ngIf="loading"></div>

            <ng-container *ngIf="!loading">
                <!-- Active orders -->
                <ng-container *ngIf="tab==='active'">
                    <div class="empty" *ngIf="activeOrders.length===0">
                        <span class="empty-emoji"></span>
                        <p>No active deliveries right now!</p>
                    </div>
                    <div class="delivery-list" *ngIf="activeOrders.length>0">
                        <div class="delivery-card card" *ngFor="let o of activeOrders">
                            <div class="flex items-center justify-between mb-2">
                                <code>{{ o.id.slice(0,8) }}…</code>
                                <span class="badge badge-{{ o.status }}">{{ 'STATUS.' + o.status | translate }}</span>
                            </div>
                            <div class="route-detail">
                                <div class="rd-row">
                                    <span class="rd-icon pickup"></span>
                                    <div>
                                        <div class="rd-label">Pickup</div>
                                        <div class="rd-addr">{{ o.pickup_address }}</div>
                                    </div>
                                </div>
                                <div class="rd-line"></div>
                                <div class="rd-row">
                                    <span class="rd-icon delivery"></span>
                                    <div>
                                        <div class="rd-label">Deliver to</div>
                                        <div class="rd-addr">{{ o.delivery_address }}</div>
                                    </div>
                                </div>
                            </div>
                            <div class="flex items-center justify-between" style="padding-top:10px;border-top:1.5px solid var(--border)">
                                <strong class="text-pink" style="font-family:var(--font-h);font-size:17px">₸{{ o.price }}</strong>
                                <span class="text-sm text-muted">{{ o.updated_at | date:'dd.MM HH:mm' }}</span>
                            </div>
                            <!-- Status actions -->
                            <div class="status-actions">
                                <ng-container *ngIf="o.status === 'assigned'">
                                    <button class="btn btn-cyan btn-sm" (click)="markInProgress(o)">
                                         Start Delivery
                                    </button>
                                    <button class="btn btn-secondary btn-sm" (click)="openHistory(o)"> History</button>
                                </ng-container>
                                <ng-container *ngIf="o.status === 'in_progress'">
                                    <button class="btn btn-success btn-sm" (click)="markDelivered(o)">
                                         Mark Delivered
                                    </button>
                                    <button class="btn btn-secondary btn-sm" (click)="openHistory(o)"> History</button>
                                </ng-container>
                            </div>
                        </div>
                    </div>
                </ng-container>

                <!-- Completed orders -->
                <ng-container *ngIf="tab==='history'">
                    <div class="empty" *ngIf="completedOrders.length===0">
                        <span class="empty-emoji"></span>
                        <p>No completed deliveries yet</p>
                    </div>
                    <div class="table-wrap" *ngIf="completedOrders.length>0">
                        <table>
                            <thead><tr>
                                <th>Order ID</th>
                                <th>Pickup</th>
                                <th>Delivery</th>
                                <th>Status</th>
                                <th>Price</th>
                                <th>Updated</th>
                                <th>History</th>
                            </tr></thead>
                            <tbody>
                            <tr *ngFor="let o of completedOrders">
                                <td><code>{{ o.id.slice(0,8) }}…</code></td>
                                <td class="addr-cell">{{ o.pickup_address }}</td>
                                <td class="addr-cell">{{ o.delivery_address }}</td>
                                <td><span class="badge badge-{{ o.status }}">{{ 'STATUS.' + o.status | translate }}</span></td>
                                <td><strong class="text-pink">₸{{ o.price }}</strong></td>
                                <td class="text-sm text-muted">{{ o.updated_at | date:'dd.MM HH:mm' }}</td>
                                <td><button class="btn btn-xs btn-secondary" (click)="openHistory(o)"></button></td>
                            </tr>
                            </tbody>
                        </table>
                    </div>
                </ng-container>

                <!-- All orders -->
                <ng-container *ngIf="tab==='all'">
                    <div class="empty" *ngIf="allOrders.length===0">
                        <span class="empty-emoji"></span>
                        <p>No orders found</p>
                    </div>
                    <div class="table-wrap" *ngIf="allOrders.length>0">
                        <table>
                            <thead><tr>
                                <th>Order ID</th>
                                <th>Pickup</th>
                                <th>Delivery</th>
                                <th>Status</th>
                                <th>Price</th>
                                <th>Actions</th>
                            </tr></thead>
                            <tbody>
                            <tr *ngFor="let o of allOrders">
                                <td><code>{{ o.id.slice(0,8) }}…</code></td>
                                <td class="addr-cell">{{ o.pickup_address }}</td>
                                <td class="addr-cell">{{ o.delivery_address }}</td>
                                <td><span class="badge badge-{{ o.status }}">{{ 'STATUS.' + o.status | translate }}</span></td>
                                <td><strong class="text-pink">₸{{ o.price }}</strong></td>
                                <td>
                                    <div class="flex gap-1">
                                        <button class="btn btn-xs btn-secondary" (click)="openHistory(o)"></button>
                                        <button *ngIf="o.status==='assigned'" class="btn btn-xs btn-cyan" (click)="markInProgress(o)"></button>
                                        <button *ngIf="o.status==='in_progress'" class="btn btn-xs btn-success" (click)="markDelivered(o)"></button>
                                    </div>
                                </td>
                            </tr>
                            </tbody>
                        </table>
                    </div>
                </ng-container>
            </ng-container>
        </div>

        <!-- Toast notification -->
        <div class="toast" *ngIf="toast" [class.toast-error]="toastIsError">{{ toast }}</div>

        <!-- History modal -->
        <div class="overlay" *ngIf="showHistory && selOrder" (click)="closeAll()">
            <div class="modal modal-wide" (click)="$event.stopPropagation()">
                <div class="modal-title"> Order History</div>
                <div class="text-sm text-muted mb-3">
                    {{ selOrder!.pickup_address }} → {{ selOrder!.delivery_address }}
                </div>
                <div class="spinner" *ngIf="loadingH"></div>
                <div class="empty" *ngIf="!loadingH && history.length===0" style="padding:20px">
                    <span style="font-size:32px"></span><p>No history yet</p>
                </div>
                <div class="timeline" *ngIf="history.length>0">
                    <div class="tl-item" *ngFor="let h of history; let i=index">
                        <div class="tl-dot" style="border-color:var(--pink);color:var(--pink)">{{ i+1 }}</div>
                        <div class="tl-content">
                            <div class="flex items-center gap-2 flex-wrap">
                                <span class="badge badge-{{ h.old_status }}">{{ h.old_status }}</span>
                                <span style="color:var(--pink);font-weight:700">→</span>
                                <span class="badge badge-{{ h.new_status }}">{{ h.new_status }}</span>
                            </div>
                            <div class="text-sm text-muted mt-1">{{ h.changed_at | date:'dd.MM.yyyy HH:mm' }}</div>
                        </div>
                    </div>
                </div>
                <button class="btn btn-secondary mt-3" (click)="closeAll()">Close</button>
            </div>
        </div>

        <!-- My courier status modal -->
        <div class="overlay" *ngIf="showMyStatus && myCourier" (click)="closeAll()">
            <div class="modal" (click)="$event.stopPropagation()">
                <div class="modal-title"> Change My Status</div>
                <div class="alert alert-success" *ngIf="modalOk">{{ modalOk }}</div>
                <div class="alert alert-error" *ngIf="modalErr">{{ modalErr }}</div>
                <div class="status-btn-grid">
                    <button class="status-btn" [class.sel]="nStatus==='free'" (click)="nStatus='free'">
                        <br><span>Free</span>
                    </button>
                    <button class="status-btn" [class.sel]="nStatus==='busy'" (click)="nStatus='busy'">
                        <br><span>Busy</span>
                    </button>
                    <button class="status-btn" [class.sel]="nStatus==='offline'" (click)="nStatus='offline'">
                        <br><span>Offline</span>
                    </button>
                </div>
                <div class="flex gap-2 mt-3">
                    <button class="btn btn-primary flex-1" (click)="saveMyStatus()" [disabled]="saving">
                        {{ saving ? '' : '' }} Save
                    </button>
                    <button class="btn btn-secondary" (click)="closeAll()">Cancel</button>
                </div>
            </div>
        </div>

        <!-- My courier transport modal -->
        <div class="overlay" *ngIf="showMyTransport && myCourier" (click)="closeAll()">
            <div class="modal" (click)="$event.stopPropagation()">
                <div class="modal-title"> Change My Transport</div>
                <div class="alert alert-success" *ngIf="modalOk">{{ modalOk }}</div>
                <div class="alert alert-error" *ngIf="modalErr">{{ modalErr }}</div>
                <div class="transport-grid">
                    <button class="transport-opt" [class.selected]="nTransport==='bike'" (click)="nTransport='bike'"><br><span>Bike</span></button>
                    <button class="transport-opt" [class.selected]="nTransport==='car'" (click)="nTransport='car'"><br><span>Car</span></button>
                    <button class="transport-opt" [class.selected]="nTransport==='foot'" (click)="nTransport='foot'"><br><span>Foot</span></button>
                    <button class="transport-opt" [class.selected]="nTransport==='scooter'" (click)="nTransport='scooter'"><br><span>Scooter</span></button>
                </div>
                <div class="flex gap-2 mt-3">
                    <button class="btn btn-primary flex-1" (click)="saveMyTransport()" [disabled]="saving">
                        {{ saving ? '' : '' }} Save
                    </button>
                    <button class="btn btn-secondary" (click)="closeAll()">Cancel</button>
                </div>
            </div>
        </div>

        <!-- My courier location modal -->
        <div class="overlay" *ngIf="showMyLocation && myCourier" (click)="closeAll()">
            <div class="modal" (click)="$event.stopPropagation()">
                <div class="modal-title"> Update My Location</div>
                <div class="alert alert-success" *ngIf="modalOk">{{ modalOk }}</div>
                <div class="alert alert-error" *ngIf="modalErr">{{ modalErr }}</div>
                <div class="grid-2">
                    <div class="form-group">
                        <label class="form-label">Latitude</label>
                        <input class="form-input" type="number" [(ngModel)]="nLat" step="0.0001">
                    </div>
                    <div class="form-group">
                        <label class="form-label">Longitude</label>
                        <input class="form-input" type="number" [(ngModel)]="nLng" step="0.0001">
                    </div>
                </div>
                <div class="mb-3">
                    <div class="form-label mb-2">Quick presets</div>
                    <div class="flex gap-2 flex-wrap">
                        <button class="btn btn-xs btn-secondary" (click)="nLat=51.1801;nLng=71.4460"> Astana</button>
                        <button class="btn btn-xs btn-secondary" (click)="nLat=43.2220;nLng=76.8512"> Almaty</button>
                        <button class="btn btn-xs btn-secondary" (click)="nLat=42.3400;nLng=69.5900"> Shymkent</button>
                    </div>
                </div>
                <div class="flex gap-2">
                    <button class="btn btn-primary flex-1" (click)="saveMyLocation()" [disabled]="saving">
                        {{ saving ? '' : '' }} Save
                    </button>
                    <button class="btn btn-secondary" (click)="closeAll()">Cancel</button>
                </div>
            </div>
        </div>
    `,
    styles: [`
        .courier-info-card { }
        .cinfo-grid { display:grid;grid-template-columns:repeat(3,1fr);gap:12px; }
        .cinfo-item { display:flex;flex-direction:column;gap:4px; }
        .cinfo-label { font-family:var(--font-h);font-size:11px;font-weight:600;text-transform:uppercase;letter-spacing:0.5px;color:var(--text2); }
        .delivery-list { display:grid;grid-template-columns:repeat(auto-fill,minmax(300px,1fr));gap:14px; }
        .delivery-card { display:flex;flex-direction:column;gap:10px; }
        .route-detail { display:flex;flex-direction:column;gap:0; }
        .rd-row { display:flex;align-items:flex-start;gap:10px; }
        .rd-icon { font-size:18px;margin-top:2px;flex-shrink:0; }
        .rd-icon.pickup { color:var(--cyan); }
        .rd-icon.delivery { color:var(--pink); }
        .rd-line { width:2px;height:16px;background:var(--border);margin-left:9px; }
        .rd-label { font-size:11px;color:var(--text2);font-family:var(--font-h);font-weight:600;text-transform:uppercase; }
        .rd-addr { font-size:13px;color:var(--text); }
        .status-actions { display:flex;gap:8px;flex-wrap:wrap;padding-top:8px;border-top:1.5px solid var(--border); }
        .addr-cell { max-width:160px;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;font-size:13px; }
        .status-btn-grid { display:grid;grid-template-columns:repeat(3,1fr);gap:8px; }
        .status-btn {
            background:var(--bg4);border:2px solid var(--border);border-radius:10px;
            padding:16px 8px;cursor:pointer;font-size:22px;text-align:center;transition:all 0.2s;
        }
        .status-btn span { display:block;font-size:12px;font-family:var(--font-h);font-weight:600;margin-top:4px;color:var(--text2); }
        .status-btn.sel { border-color:var(--pink);background:rgba(255,77,166,0.1); }
        .status-btn:hover:not(.sel) { background:var(--bg5); }
        .transport-grid { display:grid;grid-template-columns:repeat(4,1fr);gap:8px;margin-bottom:4px; }
        .transport-opt { background:var(--bg4);border:2px solid var(--border);border-radius:10px;padding:12px 6px;cursor:pointer;font-size:20px;text-align:center;transition:all 0.2s; }
        .transport-opt span { display:block;font-size:11px;margin-top:4px;font-family:var(--font-h);color:var(--text2); }
        .transport-opt.selected { border-color:var(--pink);background:rgba(255,77,166,0.1);color:var(--pink); }
        @media(max-width:768px) { .cinfo-grid{grid-template-columns:1fr 1fr;} .delivery-list{grid-template-columns:1fr;} }
    `]
})
export class CourierOrdersComponent implements OnInit {
    allOrders: Order[] = [];
    loading = false;

    tab: 'active' | 'history' | 'all' = 'active';

    myCourier: Courier | null = null;

    selOrder: Order | null = null;
    showHistory = false;
    loadingH = false;
    history: OrderHistory[] = [];

    showMyStatus = false;
    showMyTransport = false;
    showMyLocation = false;

    nStatus: 'free' | 'busy' | 'offline' = 'free';
    nTransport: 'bike' | 'car' | 'foot' | 'scooter' = 'bike';
    nLat = 0;
    nLng = 0;

    saving = false;
    modalOk = '';
    modalErr = '';

    toast = '';
    toastIsError = false;
    private toastTimer: any;

    constructor(
        private orderSvc: OrderService,
        private courierSvc: CourierService,
        public auth: AuthService
    ) {}

    ngOnInit(): void {
        this.loadAll();
    }

    get activeOrders() {
        return this.allOrders.filter(o =>
            o.status === 'assigned' || o.status === 'in_progress'
        );
    }

    get completedOrders() {
        return this.allOrders.filter(o =>
            o.status === 'delivered' || o.status === 'cancelled'
        );
    }

    loadAll(): void {
        this.loading = true;


        this.courierSvc.getMyOrders().subscribe({
            next: res => {
                this.allOrders = res ?? [];
                this.loading = false;
            },
            error: () => (this.loading = false)
        });


        this.courierSvc.getMe().subscribe({
            next: c => (this.myCourier = c),
            error: () => (this.myCourier = null)
        });
    }

    transportEmoji(t: string): string {
        return {
            bike: '',
            car: '',
            foot: '',
            scooter: ''
        }[t] ?? '';
    }

    markInProgress(o: Order): void {
        this.orderSvc.updateStatus(o.id, 'in_progress').subscribe({
            next: updated => {
                this.patchOrder(updated);
                this.showToast(' Delivery started!');
            },
            error: e => this.showToast(e.error?.error ?? 'Error', true)
        });
    }

    markDelivered(o: Order): void {
        this.orderSvc.updateStatus(o.id, 'delivered').subscribe({
            next: updated => {
                this.patchOrder(updated);
                this.showToast(' Delivered!');


                this.courierSvc.updateStatus(this.myCourier!.id, 'free').subscribe({
                    next: c => (this.myCourier = c),
                    error: () => {}
                });
            },
            error: e => this.showToast(e.error?.error ?? 'Error', true)
        });
    }

    openHistory(o: Order): void {
        this.selOrder = o;
        this.history = [];
        this.loadingH = true;
        this.showHistory = true;

        this.orderSvc.getHistory(o.id).subscribe({
            next: h => {
                this.history = h ?? [];
                this.loadingH = false;
            },
            error: () => (this.loadingH = false)
        });
    }


    openMyStatus(): void {
        this.resetModal();
        if (this.myCourier) this.nStatus = this.myCourier.status;
        this.showMyStatus = true;
    }

    openMyTransport(): void {
        this.resetModal();
        if (this.myCourier) this.nTransport = this.myCourier.transport_type;
        this.showMyTransport = true;
    }

    openMyLocation(): void {
        this.resetModal();
        if (this.myCourier) {
            this.nLat = this.myCourier.current_lat;
            this.nLng = this.myCourier.current_lng;
        }
        this.showMyLocation = true;
    }

    closeAll(): void {
        this.showHistory = false;
        this.showMyStatus = false;
        this.showMyTransport = false;
        this.showMyLocation = false;
        this.selOrder = null;
    }

    private resetModal(): void {
        this.modalOk = '';
        this.modalErr = '';
    }


    saveMyStatus(): void {
        if (!this.myCourier) return;

        this.saving = true;
        this.modalErr = '';

        this.courierSvc.updateStatus(this.myCourier!.id, this.nStatus).subscribe({
            next: c => {
                this.myCourier = c;
                this.successAndClose('Status updated!');
            },
            error: e => this.fail(e)
        });
    }

    saveMyTransport(): void {
        if (!this.myCourier) return;

        this.saving = true;
        this.modalErr = '';

        this.courierSvc.updateTransport(this.myCourier.id, this.nTransport).subscribe({
            next: c => {
                this.myCourier = c;
                this.successAndClose('Transport updated!');
            },
            error: e => this.fail(e)
        });
    }

    saveMyLocation(): void {
        if (!this.myCourier) return;

        this.saving = true;
        this.modalErr = '';

        this.courierSvc.updateLocation(
            this.myCourier.id,
            this.nLat,
            this.nLng
        ).subscribe({
            next: c => {
                this.myCourier = c;
                this.successAndClose('Location updated!');
            },
            error: e => this.fail(e)
        });
    }


    private successAndClose(msg: string): void {
        this.modalOk = msg;
        this.saving = false;

        setTimeout(() => this.closeAll(), 1000);
    }

    private fail(e: any): void {
        this.modalErr = e.error?.error ?? 'Error';
        this.saving = false;
    }

    private patchOrder(updated: Order): void {
        const i = this.allOrders.findIndex(o => o.id === updated.id);
        if (i >= 0) this.allOrders[i] = updated;
    }

    private showToast(msg: string, error = false): void {
        this.toast = msg;
        this.toastIsError = error;

        clearTimeout(this.toastTimer);
        this.toastTimer = setTimeout(() => (this.toast = ''), 3000);
    }
}