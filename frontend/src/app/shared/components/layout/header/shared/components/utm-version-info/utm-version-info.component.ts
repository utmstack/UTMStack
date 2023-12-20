import {Component, OnDestroy, OnInit} from '@angular/core';
import {CheckForUpdatesService} from '../../../../../../services/updates/check-for-updates.service';
import {VersionInfo} from '../../../../../../types/updates/updates.type';

@Component({
  selector: 'app-utm-version-info',
  templateUrl: './utm-version-info.component.html',
  styleUrls: ['./utm-version-info.component.css']
})
export class UtmVersionInfoComponent implements OnInit, OnDestroy {
  currentVersion: VersionInfo;

  constructor(private checkForUpdatesService: CheckForUpdatesService) {
  }

  ngOnInit() {
    this.getVersionInfo();
  }

  ngOnDestroy() {
  }

  getVersionInfo() {
    this.checkForUpdatesService.getVersion().subscribe(response => {
      this.currentVersion = response.body;
    });
  }

}
