import {HttpClient} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {BehaviorSubject, Observable} from 'rxjs';
import {SERVER_API_URL} from '../../app.constants';
import {DATE_SETTING_FORMAT_SHORT, DATE_SETTING_TIMEZONE_SHORT} from '../constants/date-timezone-date.const';
import {DatePipeDefaultOptions} from '../types/date-pipe-default-options';

@Injectable({
  providedIn: 'root'
})
export class TimezoneFormatService {
  public resourceUrl = SERVER_API_URL + 'api/date-format';
  private dateFormatSubject: BehaviorSubject<DatePipeDefaultOptions> = new BehaviorSubject<DatePipeDefaultOptions>({
    dateFormat: 'medium',
    timezone: 'UTC'
  });

  constructor(private httpClient: HttpClient) {
  }

  loadTimezoneAndFormat(): void {
    this.httpClient.get(this.resourceUrl)
      .subscribe(response => {
        const configs = response;
        const timezone = this.getSettingValue(configs, DATE_SETTING_TIMEZONE_SHORT);
        const dateFormat = this.getSettingValue(configs, DATE_SETTING_FORMAT_SHORT);
        this.dateFormatSubject.next({dateFormat: dateFormat ? dateFormat : 'medium', timezone: timezone ? timezone : 'UTC'});
      }, error => {
        console.error('Unable to set default time format');
      });
  }

  getSettingValue(settings: any, key: string): string {
    if (settings) {
      return settings[key];
    } else {
      return null;
    }
  }

  getDateFormatSubject(): Observable<DatePipeDefaultOptions> {
    return this.dateFormatSubject.asObservable();
  }

  setDateFormatSubject(format: DatePipeDefaultOptions) {
    console.log('Updating app date settings', format);
    return this.dateFormatSubject.next(format);
  }

}
