import {Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';

@Injectable({providedIn: 'root'})
export class ThemeChangeBehavior {
  $themeChange = new BehaviorSubject<boolean>(false);
  $themeReportIcon = new BehaviorSubject<string>(null);
  $themeReportCover = new BehaviorSubject<string>(null);
  $themeIcon = new BehaviorSubject<string>(null);
  $themeNavbarIcon = new BehaviorSubject<string>(null);
}
