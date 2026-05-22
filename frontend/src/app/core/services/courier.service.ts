import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../../environments/environment';
import { Courier, Order } from '../../shared/models/models';
import { AuthService } from './auth.service';

@Injectable({ providedIn: 'root' })
export class CourierService {
    private apiUrl = environment.apiUrl;

    constructor(private http: HttpClient, private auth: AuthService) {}

    listCouriers(): Observable<Courier[]> {
        return this.http.get<Courier[]>(`${this.apiUrl}/couriers`);
    }

    listFreeCouriers(): Observable<Courier[]> {
        return this.http.get<Courier[]>(`${this.apiUrl}/couriers/free`);
    }

    updateStatus(id: string, status: string): Observable<Courier> {
        const url = this.auth.role === 'courier'
            ? `${this.apiUrl}/couriers/me/status`
            : `${this.apiUrl}/couriers/${id}/status`;
        return this.http.patch<Courier>(url, { status });
    }

    updateTransport(id: string, transport_type: string): Observable<Courier> {
        const url = this.auth.role === 'courier'
            ? `${this.apiUrl}/couriers/me/transport`
            : `${this.apiUrl}/couriers/${id}/transport`;
        return this.http.patch<Courier>(url, { transport_type });
    }

    updateLocation(id: string, lat: number, lng: number): Observable<Courier> {
        const url = this.auth.role === 'courier'
            ? `${this.apiUrl}/couriers/me/location`
            : `${this.apiUrl}/couriers/${id}/location`;
        return this.http.patch<Courier>(url, { lat, lng });
    }
    getMe(): Observable<Courier> {
        return this.http.get<Courier>(`${this.apiUrl}/couriers/me`);
    }

    getMyOrders(): Observable<Order[]> {
        return this.http.get<Order[]>(`${this.apiUrl}/couriers/me/orders`);
    }
}