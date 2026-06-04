import {HttpClient} from '@angular/common/http';
import {Injectable, NgZone} from '@angular/core';
import {BehaviorSubject, Observable} from 'rxjs';
import {SERVER_API_URL} from '../../app.constants';
import {AuthServerProvider} from '../../core/auth/auth-jwt.service';

// import {EventSourcePolyfill} from "ng-event-source";

@Injectable({providedIn: 'root'})
export class UtmWebfluxService {
  serverApiUrl = SERVER_API_URL;
  $events: BehaviorSubject<any> = new BehaviorSubject<any>(null);
  $connected: BehaviorSubject<boolean> = new BehaviorSubject<boolean>(true);
  source: EventSource;

  constructor(private $http: HttpClient,
              private $zone: NgZone,
              private authServerProvider: AuthServerProvider) {
  }

  public createEventObservable(): Observable<any> {
    return this.$events.asObservable();
  }

  public createConnectionObservable(): Observable<any> {
    return this.$connected.asObservable();
  }

  public createEventSource(): EventSource {
    const authToken = this.authServerProvider.getToken();
    let url = this.serverApiUrl + 'api/ad/utm-ad-trackers/tracking';
    // let url = this.serverApiUrl + 'api/ad/utm-ad-trackers/numeros';
    if (authToken) {
      url += '?access_token=' + authToken;
    }
    const eventSource = new EventSource(url, {withCredentials: true});
    this.source = eventSource;
    eventSource.onmessage = sse => {
      this.$zone.run(() => {
        this.$events.next(sse.data);
        this.$connected.next(true);
      });
    };
    eventSource.onerror = err => {
      switch (eventSource.readyState) {
        case eventSource.CONNECTING:
          // console.log('CONNECTED TO NOTIFICATION...');
          break;
        case eventSource.CLOSED:
          this.$connected.next(false);
          setInterval(() => {
            this.createEventSource();
          }, 3000);
          break;
      }
    };
    return eventSource;
  }

  close() {
    this.source.close();
    // this.$events.unsubscribe();
  }
}
