import {Component, Input, OnInit} from '@angular/core';
import {UtmToastService} from '../../../../../../alert/utm-toast.service';
import {AppVersionService} from '../../../../../../services/version/app-version.service';
import {AppVersionInfo} from '../../../../../../types/updates/updates.type';

@Component({
  selector: 'app-utm-version-info',
  templateUrl: './utm-version-info.component.html',
  styleUrls: ['./utm-version-info.component.css']
})
export class UtmVersionInfoComponent {
 @Input('version') versionInfo: AppVersionInfo;

  constructor() {}
}
