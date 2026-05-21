import { Injectable } from '@angular/core';
import {
  HttpRequest, HttpHandler, HttpEvent,
  HttpInterceptor, HttpErrorResponse
} from '@angular/common/http';
import { Observable, throwError, catchError, switchMap } from 'rxjs';
import { AuthService } from '../services/auth.service';
import { Router } from '@angular/router';

@Injectable()
export class AuthInterceptor implements HttpInterceptor {
  constructor(private auth: AuthService, private router: Router) {}

  intercept(req: HttpRequest<unknown>, next: HttpHandler): Observable<HttpEvent<unknown>> {
    const token = localStorage.getItem('access_token');
    const authReq = token
      ? req.clone({ setHeaders: { Authorization: `Bearer ${token}` } })
      : req;

    return next.handle(authReq).pipe(
      catchError((err: HttpErrorResponse) => {
        if (err.status === 401 && !req.url.includes('/auth/')) {
          return this.auth.refresh().pipe(
            switchMap(() => {
              const newToken = localStorage.getItem('access_token')!;
              const retried = req.clone({ setHeaders: { Authorization: `Bearer ${newToken}` } });
              return next.handle(retried);
            }),
            catchError(() => {
              this.auth.logout();
              return throwError(() => err);
            })
          );
        }
        return throwError(() => err);
      })
    );
  }
}
