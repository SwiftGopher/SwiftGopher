import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { BehaviorSubject, Observable, tap } from 'rxjs';
import { Router } from '@angular/router';
import { environment } from '../../../environments/environment';
import { User, TokenPair, LoginRequest, RegisterRequest } from '../../shared/models/models';

@Injectable({ providedIn: 'root' })
export class AuthService {
  private apiUrl = environment.apiUrl;
  private _user = new BehaviorSubject<User | null>(null);
  user$ = this._user.asObservable();

  constructor(private http: HttpClient, private router: Router) {
    this.loadUserFromStorage();
  }

  get currentUser(): User | null {
    return this._user.value;
  }

  get isLoggedIn(): boolean {
    return !!localStorage.getItem('access_token');
  }

  get role(): string {
    return this._user.value?.role ?? '';
  }

  login(req: LoginRequest): Observable<TokenPair> {
    return this.http.post<TokenPair>(`${this.apiUrl}/auth/login`, req).pipe(
      tap(tokens => {
        localStorage.setItem('access_token', tokens.access_token);
        localStorage.setItem('refresh_token', tokens.refresh_token);
        this.fetchProfile();
      })
    );
  }

  register(req: RegisterRequest): Observable<User> {
    return this.http.post<User>(`${this.apiUrl}/auth/register`, req);
  }

  logout(): void {
    const token = localStorage.getItem('access_token');
    if (token) {
      this.http.post(`${this.apiUrl}/auth/logout`, {}).subscribe();
    }
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    this._user.next(null);
    this.router.navigate(['/login']);
  }

  fetchProfile(): void {
    this.http.get<User>(`${this.apiUrl}/profile`).subscribe({
      next: u => this._user.next(u),
      error: () => {}
    });
  }

  refresh(): Observable<TokenPair> {
    const refresh_token = localStorage.getItem('refresh_token') ?? '';
    return this.http.post<TokenPair>(`${this.apiUrl}/auth/refresh`, { refresh_token }).pipe(
      tap(tokens => {
        localStorage.setItem('access_token', tokens.access_token);
        localStorage.setItem('refresh_token', tokens.refresh_token);
      })
    );
  }

  private loadUserFromStorage(): void {
    if (this.isLoggedIn) {
      this.fetchProfile();
    }
  }
}
