// Vendored from ng-jhipster (JHipster), Apache-2.0 licensed.
//
// ng-jhipster@0.16.0 is the last published release and pins Angular ^10 peers,
// which blocks the Angular 20 migration (it fails `ng update`'s peer resolution
// and has no modern fork). The app imports exactly five of its utilities
// (JhiParseLinks, JhiResolvePagingParams, JhiEventManager, JhiAlertService,
// JhiBase64Service); those plus their three transitive deps are reproduced here,
// so the `ng-jhipster` package can be dropped entirely.
//
// All classes are `providedIn: 'root'` injectables using only standard
// @angular/* and rxjs APIs — identical behavior to the original.
import {Injectable, NgZone, Optional, SecurityContext} from '@angular/core';
import {DomSanitizer} from '@angular/platform-browser';
import {Observable, Subscription} from 'rxjs';
import {filter, map, share} from 'rxjs/operators';
import {TranslateService} from '@ngx-translate/core';
import {faSort, faSortDown, faSortUp} from '@fortawesome/free-solid-svg-icons';

@Injectable({providedIn: 'root'})
export class JhiModuleConfig {
  sortIcon = faSort;
  sortAscIcon = faSortUp;
  sortDescIcon = faSortDown;
  i18nEnabled = false;
  defaultI18nLang = 'en';
  noi18nMessage = 'translation-not-found';
  alertAsToast = false;
  alertTimeout = 5000;
  classBadgeTrue = 'badge badge-success';
  classBadgeFalse = 'badge badge-danger';
  classTrue = 'fa fa-lg fa-check text-success';
  classFalse = 'fa fa-lg fa-times text-danger';
}

@Injectable({providedIn: 'root'})
export class JhiConfigService {
  CONFIG_OPTIONS: any;
  constructor(moduleConfig: JhiModuleConfig) {
    this.CONFIG_OPTIONS = Object.assign(Object.assign({}, new JhiModuleConfig()), moduleConfig);
  }

  getConfig() {
    return this.CONFIG_OPTIONS;
  }
}

/** Utility for parsing the `Link` header of paginated REST responses. */
@Injectable({providedIn: 'root'})
export class JhiParseLinks {
  parse(header: string): any {
    if (header.length === 0) {
      throw new Error('input must not be of zero length');
    }
    const parts = header.split(',');
    const links: {[rel: string]: number} = {};
    parts.forEach(p => {
      const section = p.split(';');
      if (section.length !== 2) {
        throw new Error('section could not be split on ";"');
      }
      const url = section[0].replace(/<(.*)>/, '$1').trim();
      const queryString: {[key: string]: any} = {};
      url.replace(new RegExp('([^?=&]+)(=([^&]*))?', 'g'), ($0, $1, $2, $3) => (queryString[$1] = $3));
      let page = queryString.page;
      if (typeof page === 'string') {
        page = parseInt(page, 10);
      }
      const name = section[1].replace(/rel="(.*)"/, '$1').trim();
      links[name] = page;
    });
    return links;
  }
}

/** Shared Subject-based event bus for cross-component broadcasts. */
@Injectable({providedIn: 'root'})
export class JhiEventManager {
  private observer: any;
  readonly observable: Observable<any>;

  constructor() {
    this.observable = new Observable(observer => {
      this.observer = observer;
    }).pipe(share());
  }

  broadcast(event: any) {
    if (this.observer) {
      this.observer.next(event);
    }
  }

  subscribe(eventName: string, callback: (event: any) => void): Subscription {
    const subscriber = this.observable
      .pipe(
        filter((event) => {
          if (typeof event === 'string') {
            return event === eventName;
          }
          return event.name === eventName;
        }),
        map((event) => {
          if (typeof event !== 'string') {
            return event;
          }
        })
      )
      .subscribe(callback);
    return subscriber;
  }

  destroy(subscriber: Subscription) {
    subscriber.unsubscribe();
  }
}

/** Base64 encode/decode for file uploads. */
@Injectable({providedIn: 'root'})
export class JhiBase64Service {
  private keyStr = 'ABCDEFGHIJKLMNOP' + 'QRSTUVWXYZabcdef' + 'ghijklmnopqrstuv' + 'wxyz0123456789+/' + '=';

  encode(input: string): string {
    let output = '';
    let i = 0;
    while (i < input.length) {
      const chr1 = input.charCodeAt(i++);
      const chr2 = input.charCodeAt(i++);
      const chr3 = input.charCodeAt(i++);
      let enc1 = chr1 >> 2;
      let enc2 = ((chr1 & 3) << 4) | (chr2 >> 4);
      let enc3 = ((chr2 & 15) << 2) | (chr3 >> 6);
      let enc4 = chr3 & 63;
      if (isNaN(chr2)) {
        enc3 = enc4 = 64;
      } else if (isNaN(chr3)) {
        enc4 = 64;
      }
      output += this.keyStr.charAt(enc1) + this.keyStr.charAt(enc2) + this.keyStr.charAt(enc3) + this.keyStr.charAt(enc4);
    }
    return output;
  }

  decode(input: string): string {
    let output = '';
    input = input.replace(/[^A-Za-z0-9+/=]/g, '');
    let i = 0;
    while (i < input.length) {
      const enc1 = this.keyStr.indexOf(input.charAt(i++));
      const enc2 = this.keyStr.indexOf(input.charAt(i++));
      const enc3 = this.keyStr.indexOf(input.charAt(i++));
      const enc4 = this.keyStr.indexOf(input.charAt(i++));
      const chr1 = (enc1 << 2) | (enc2 >> 4);
      const chr2 = ((enc2 & 15) << 4) | (enc3 >> 2);
      const chr3 = ((enc3 & 3) << 6) | enc4;
      output += String.fromCharCode(chr1);
      if (enc3 !== 64) {
        output += String.fromCharCode(chr2);
      }
      if (enc4 !== 64) {
        output += String.fromCharCode(chr3);
      }
    }
    return output;
  }
}

/** Alert service for success/error/warning/info toasts (JHipster-style). */
@Injectable({providedIn: 'root'})
export class JhiAlertService {
  toast: boolean;
  i18nEnabled: boolean;
  private alertId = 0;
  private alerts: any[] = [];
  timeout: number;


  constructor(
    private sanitizer: DomSanitizer,
    configService: JhiConfigService,
    private ngZone: NgZone,
    @Optional() private translateService: TranslateService
  ) {
    const config = configService.getConfig();
    this.toast = config.alertAsToast;
    this.i18nEnabled = config.i18nEnabled;
    this.timeout = config.alertTimeout;
  }

  clear() {
    this.alerts.splice(0, this.alerts.length);
  }

  get() {
    return this.alerts;
  }

  success(msg: string, params?: any, position?: string) {
    return this.addAlert({type: 'success', msg, params, timeout: this.timeout, toast: this.isToast(), position}, []);
  }

  error(msg: string, params?: any, position?: string) {
    return this.addAlert({type: 'danger', msg, params, timeout: this.timeout, toast: this.isToast(), position}, []);
  }

  warning(msg: string, params?: any, position?: string) {
    return this.addAlert({type: 'warning', msg, params, timeout: this.timeout, toast: this.isToast(), position}, []);
  }

  info(msg: string, params?: any, position?: string) {
    return this.addAlert({type: 'info', msg, params, timeout: this.timeout, toast: this.isToast(), position}, []);
  }

  addAlert(alertOptions: any, extAlerts: any[]) {
    alertOptions.id = this.alertId++;
    if (this.i18nEnabled && alertOptions.msg) {
      alertOptions.msg = this.translateService.instant(alertOptions.msg, alertOptions.params);
    }
    const alert = this.factory(alertOptions);
    if (alertOptions.timeout && alertOptions.timeout > 0) {
      // Workaround protractor waiting for setTimeout.
      this.ngZone.runOutsideAngular(() => {
        setTimeout(() => {
          this.ngZone.run(() => {
            this.closeAlert(alertOptions.id, extAlerts);
          });
        }, alertOptions.timeout);
      });
    }
    return alert;
  }

  closeAlert(id: number, extAlerts: any[]) {
    const thisAlerts = extAlerts && extAlerts.length > 0 ? extAlerts : this.alerts;
    return this.closeAlertByIndex(thisAlerts.map(e => e.id).indexOf(id), thisAlerts);
  }

  closeAlertByIndex(index: number, thisAlerts: any[]) {
    return thisAlerts.splice(index, 1);
  }

  isToast() {
    return this.toast;
  }

  factory(alertOptions: any) {
    const alert = {
      type: alertOptions.type,
      msg: this.sanitizer.sanitize(SecurityContext.HTML, alertOptions.msg),
      id: alertOptions.id,
      timeout: alertOptions.timeout,
      toast: alertOptions.toast,
      position: alertOptions.position ? alertOptions.position : 'top right',
      scoped: alertOptions.scoped,
      close: (alerts: any[]) => this.closeAlert(alertOptions.id, alerts)
    };
    if (!alert.scoped) {
      this.alerts.push(alert);
    }
    return alert;
  }
}

/** Pure helpers for parsing `page`/`sort` query params. */
@Injectable({providedIn: 'root'})
export class JhiPaginationUtil {
  parseAscending(sort: string) {
    let sortArray = sort.split(',');
    sortArray = sortArray.length > 1 ? sortArray : sort.split('%2C');
    if (sortArray.length > 1) {
      return sortArray.slice(-1)[0] === 'asc';
    }
    // default to true if no sort is defined
    return true;
  }

  parsePage(page: string) {
    return parseInt(page, 10);
  }

  parsePredicate(sort: string) {
    return sort.split(',')[0].split('%2C')[0];
  }
}

/** Route resolver that produces {page, predicate, ascending} from query params. */
@Injectable({providedIn: 'root'})
export class JhiResolvePagingParams {
  constructor(private paginationUtil: JhiPaginationUtil) {}

  resolve(route: any, state: any) {
    const page = route.queryParams['page'] ? route.queryParams['page'] : '1';
    const defaultSort = route.data['defaultSort'] ? route.data['defaultSort'] : 'id,asc';
    const sort = route.queryParams['sort'] ? route.queryParams['sort'] : defaultSort;
    return {
      page: this.paginationUtil.parsePage(page),
      predicate: this.paginationUtil.parsePredicate(sort),
      ascending: this.paginationUtil.parseAscending(sort)
    };
  }
}
