import {Injectable} from '@angular/core';
import {NavigationEnd, Router} from '@angular/router';
import {Observable, Observer, Subscription} from 'rxjs';

import * as SockJS from 'sockjs-client';
import * as Stomp from 'webstomp-client';
import {WS_SERVER_API_URL} from '../../app.constants';

import {AuthServerProvider} from '../auth/auth-jwt.service';
import {WindowRef} from './window.service';

@Injectable({providedIn: 'root'})
export class UtmTrackerService {
  stompClient = null;
  subscriber = null;
  connection: Promise<any>;
  connectedPromise: any;
  listener: Observable<any>;
  listenerObserver: Observer<any>;
  alreadyConnectedOnce = false;
  private subscription: Subscription;

  constructor(
    private router: Router,
    private authServerProvider: AuthServerProvider,
    private $window: WindowRef,
  ) {
    this.connection = this.createConnection();
    this.listener = this.createListener();
  }

  connect() {
    if (this.connectedPromise === null) {
      this.connection = this.createConnection();
    }
    // building absolute path so that websocket doesn't fail when deploying with a context path
    const loc = this.$window.nativeWindow.location;
    let url;
    url = WS_SERVER_API_URL + '/tracker';
    const authToken = this.authServerProvider.getToken();
    if (authToken) {
      url += '?access_token=' + authToken;
    }
    const socket = new SockJS(url);
    this.stompClient = Stomp.over(socket);
    const headers = {};
    this.stompClient.connect(
      headers,
      () => {
        this.connectedPromise('success');
        this.connectedPromise = null;
        this.sendActivity();
        if (!this.alreadyConnectedOnce) {
          this.subscription = this.router.events.subscribe(event => {
            if (event instanceof NavigationEnd) {
              this.sendActivity();
            }
          });
          this.alreadyConnectedOnce = true;
        }
      }
    );
  }

  disconnect() {
    if (this.stompClient !== null) {
      this.stompClient.disconnect();
      this.stompClient = null;
    }
    if (this.subscription) {
      this.subscription.unsubscribe();
      this.subscription = null;
    }
    this.alreadyConnectedOnce = false;
  }

  receive() {
    return this.listener;
  }

  sendActivity() {
    if (this.stompClient !== null && this.stompClient.connected) {
      this.stompClient.send(
        '/topic/activity',
        JSON.stringify({page: this.router.routerState.snapshot.url}), // body
        {} // header
      );
    }
  }

  subscribe() {
    this.connection.then(() => {
      this.subscriber = this.stompClient.subscribe('/topic/tracker', data => {
        this.listenerObserver.next(JSON.parse(data.body));
      });
    });
  }

  unsubscribe() {
    if (this.subscriber !== null) {
      this.subscriber.unsubscribe();
    }
    this.listener = this.createListener();
  }

  private createListener(): Observable<any> {
    return new Observable(observer => {
      this.listenerObserver = observer;
    });
  }

  private createConnection(): Promise<any> {
    return new Promise((resolve, reject) => (this.connectedPromise = resolve));
  }
}
