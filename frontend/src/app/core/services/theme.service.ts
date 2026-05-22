import { Injectable } from '@angular/core';
import { BehaviorSubject } from 'rxjs';

@Injectable({ providedIn: 'root' })
export class ThemeService {
  private _theme = new BehaviorSubject<'dark' | 'light'>('dark');
  theme$ = this._theme.asObservable();

  constructor() {
    const saved = localStorage.getItem('theme') as 'dark' | 'light' | null;
    const theme = saved ?? 'dark';
    this.apply(theme);
  }

  toggle(): void {
    this.apply(this._theme.value === 'dark' ? 'light' : 'dark');
  }

  get current(): 'dark' | 'light' {
    return this._theme.value;
  }

  private apply(theme: 'dark' | 'light'): void {
    document.documentElement.setAttribute('data-theme', theme);
    localStorage.setItem('theme', theme);
    this._theme.next(theme);
  }
}
