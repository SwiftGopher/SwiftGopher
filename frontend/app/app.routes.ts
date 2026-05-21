import { Routes } from '@angular/router';
import { authGuard } from './core/guards/auth.guard';

export const routes: Routes = [
  { path: '', redirectTo: 'dashboard', pathMatch: 'full' },
  {
    path: 'login',
    loadComponent: () => import('./features/auth/login.component').then(m => m.LoginComponent)
  },
  {
    path: 'register',
    loadComponent: () => import('./features/auth/register.component').then(m => m.RegisterComponent)
  },
  {
    path: 'dashboard',
    canActivate: [authGuard],
    loadComponent: () => import('./features/dashboard/dashboard.component').then(m => m.DashboardComponent)
  },
  {
    path: 'orders',
    canActivate: [authGuard],
    loadComponent: () => import('./features/orders/orders.component').then(m => m.OrdersComponent)
  },
  {
    path: 'my-orders',
    canActivate: [authGuard],
    loadComponent: () => import('./features/orders/my-orders.component').then(m => m.MyOrdersComponent)
  },
  {
    path: 'couriers',
    canActivate: [authGuard],
    loadComponent: () => import('./features/couriers/couriers.component').then(m => m.CouriersComponent)
  },
  {
    path: 'my-deliveries',
    canActivate: [authGuard],
    loadComponent: () => import('./features/courier-orders/courier-orders.component').then(m => m.CourierOrdersComponent)
  },
  {
    path: 'profile',
    canActivate: [authGuard],
    loadComponent: () => import('./features/profile/profile.component').then(m => m.ProfileComponent)
  },
  { path: '**', redirectTo: 'dashboard' }
];
