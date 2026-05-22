import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../../environments/environment';
import { Order, OrderHistory, OrderFilter, CreateOrderRequest, OrderStatus } from '../../shared/models/models';

@Injectable({ providedIn: 'root' })
export class OrderService {
  private apiUrl = environment.apiUrl;

  constructor(private http: HttpClient) {}

  createOrder(req: CreateOrderRequest): Observable<Order> {
    return this.http.post<Order>(`${this.apiUrl}/orders`, req);
  }

  getOrder(id: string): Observable<Order> {
    return this.http.get<Order>(`${this.apiUrl}/orders/${id}`);
  }

  listOrders(filter: OrderFilter = {}): Observable<{ data: Order[]; limit: number; offset: number }> {
    let params = new HttpParams();
    if (filter.status) params = params.set('status', filter.status);
    if (filter.limit !== undefined)  params = params.set('limit', String(filter.limit));
    if (filter.offset !== undefined) params = params.set('offset', String(filter.offset));
    if (filter.sort_by)  params = params.set('sort_by', filter.sort_by);
    if (filter.sort_dir) params = params.set('sort_dir', filter.sort_dir);
    return this.http.get<{ data: Order[]; limit: number; offset: number }>(
      `${this.apiUrl}/orders`, { params }
    );
  }

  getMyOrders(): Observable<Order[]> {
    return this.http.get<Order[]>(`${this.apiUrl}/orders/my`);
  }

  updateStatus(id: string, status: OrderStatus): Observable<Order> {
    return this.http.patch<Order>(`${this.apiUrl}/orders/${id}/status`, { status });
  }

  getHistory(id: string): Observable<OrderHistory[]> {
    return this.http.get<OrderHistory[]>(`${this.apiUrl}/orders/${id}/history`);
  }
  getMyCourierOrders(): Observable<Order[]> {
    return this.http.get<Order[]>(
        `${this.apiUrl}/couriers/me/orders`
    );
  }
}
