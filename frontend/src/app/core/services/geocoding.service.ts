import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, of } from 'rxjs';
import { map, catchError } from 'rxjs/operators';

export interface GeocodingResult {
    lat: number;
    lng: number;
    displayName: string;
}

@Injectable({ providedIn: 'root' })
export class GeocodingService {
    private readonly nominatimUrl = 'https://nominatim.openstreetmap.org/search';

    constructor(private http: HttpClient) {}

    geocode(address: string): Observable<GeocodingResult | null> {
        if (!address || address.trim().length < 3) {
            return of(null);
        }

        const params = {
            q: address,
            format: 'json',
            limit: '1',
            countrycodes: 'kz',  
            addressdetails: '1',
        };

        return this.http.get<any[]>(this.nominatimUrl, { params }).pipe(
            map(results => {
                if (!results || results.length === 0) return null;
                const r = results[0];
                return {
                    lat: parseFloat(r.lat),
                    lng: parseFloat(r.lon),
                    displayName: r.display_name,
                };
            }),
            catchError(() => of(null))
        );
    }
}