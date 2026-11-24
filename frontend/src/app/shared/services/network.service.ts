import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import {BehaviorSubject, Observable, of, Subscription, timer} from 'rxjs';
import { catchError, map, switchMap } from 'rxjs/operators';

@Injectable({
  providedIn: 'root',
})
export class NetworkService {
  private isOnlineSubject: BehaviorSubject<boolean> = new BehaviorSubject<boolean>(navigator.onLine);
  isOnline$: Observable<boolean> = this.isOnlineSubject.asObservable();
  private connectivityCheckInterval = 60000;
  private connectivityCheckUrl = '/api/info/connectivity';
  private connectivityCheckSubscription: Subscription | undefined;

  constructor(private http: HttpClient) {
    window.addEventListener('online', this.updateOnlineStatus.bind(this));
    window.addEventListener('offline', this.updateOnlineStatus.bind(this));
    this.startConnectivityCheck();
  }

  private updateOnlineStatus(): void {
    this.isOnlineSubject.next(navigator.onLine);
    if (navigator.onLine) {
      this.startConnectivityCheck();
    } else {
      this.stopConnectivityCheck();
    }
  }

  private startConnectivityCheck(): void {
    if (!this.connectivityCheckSubscription) {
      this.connectivityCheckSubscription = timer(0, this.connectivityCheckInterval).pipe(
        switchMap(() => this.checkInternetConnectivity())
      ).subscribe(isConnected => {
        if (this.isOnlineSubject.value !== isConnected) {
          this.isOnlineSubject.next(isConnected);
        }
      });
    }
  }

  private stopConnectivityCheck(): void {
    if (this.connectivityCheckSubscription) {
      this.connectivityCheckSubscription.unsubscribe();
      this.connectivityCheckSubscription = undefined;
    }
  }

  private checkInternetConnectivity(): Observable<boolean> {
    return this.http.get<{ online: boolean }>(this.connectivityCheckUrl).pipe(
      map(response => Boolean(response?.online)),
      catchError(() => of(false))
    );
  }

  get isOnline(): boolean {
    return this.isOnlineSubject.getValue();
  }

  getOnlineStatus(): Observable<boolean> {
    return this.isOnline$;
  }

  setConnectivityCheckUrl(url: string): void {
    this.connectivityCheckUrl = url;
    this.stopConnectivityCheck();
    this.startConnectivityCheck();
  }

  setConnectivityCheckInterval(interval: number): void {
    this.connectivityCheckInterval = interval;
    this.stopConnectivityCheck();
    this.startConnectivityCheck();
  }
}
