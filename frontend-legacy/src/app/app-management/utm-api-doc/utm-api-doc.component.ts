import {Component, OnDestroy, OnInit} from '@angular/core';
import {Subject} from 'rxjs';
import {takeUntil} from 'rxjs/operators';
import {VersionInfoService} from '../../shared/services/version/version-info.service';
import {AppVersionInfo} from '../../shared/types/updates/updates.type';

@Component({
  selector: 'app-utm-api-doc',
  templateUrl: './utm-api-doc.component.html',
  styleUrls: ['./utm-api-doc.component.scss']
})
export class UtmApiDocComponent implements OnInit, OnDestroy {

  versionInfo: AppVersionInfo;
  destroy$ = new Subject<void>();

  constructor(private versionTypeService: VersionInfoService) {
  }

  ngOnInit() {

    this.versionTypeService.appVersionInfo$
      .pipe(takeUntil(this.destroy$))
      .subscribe((versionInfo: AppVersionInfo) => this.versionInfo = versionInfo);
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

}
