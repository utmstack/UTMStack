import {Component, OnDestroy, OnInit} from '@angular/core';
import {EMPTY, Observable, Subject} from 'rxjs';
import {catchError, map, tap} from 'rxjs/operators';
import {UtmToastService} from '../../../../../../alert/utm-toast.service';
import {CheckForUpdatesService} from '../../../../../../services/updates/check-for-updates.service';
import {VersionType, VersionTypeService} from '../../../../../../services/util/version-type.service';
import {VersionInfo} from '../../../../../../types/updates/updates.type';

@Component({
  selector: 'app-utm-version-info',
  templateUrl: './utm-version-info.component.html',
  styleUrls: ['./utm-version-info.component.css']
})
export class UtmVersionInfoComponent implements OnInit {
  versionInfo: VersionInfo;

  constructor(private checkForUpdatesService: CheckForUpdatesService,
              private utmToastService: UtmToastService,
              private versionTypeService: VersionTypeService) {
  }

  ngOnInit() {
    this.checkForUpdatesService.getVersion()
      .pipe(
        map(response => response.body || null),
        tap((versionInfo: VersionInfo) => {
          const version = versionInfo && versionInfo.build && versionInfo.build.version || '';
          const versionType = version.includes('community') || version === ''
            ? VersionType.COMMUNITY
            : VersionType.ENTERPRISE;

          if (versionType !== this.versionTypeService.versionType()) {
            this.versionTypeService.changeVersionType(versionType);
          }
        }),
        catchError(() => {
          this.utmToastService.showError(
            'Error fetching version info',
            'An error occurred while fetching version info.'
          );
          return EMPTY;
        })
      )
      .subscribe(versionInfo => {
        this.versionInfo = versionInfo;
      });
  }

}
