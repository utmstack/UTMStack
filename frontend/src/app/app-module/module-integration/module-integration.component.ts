import {Component, Input, OnInit} from '@angular/core';
import {CheckForUpdatesService} from '../../shared/services/updates/check-for-updates.service';
import {UtmModulesEnum} from '../shared/enum/utm-module.enum';
import {UtmModuleType} from '../shared/type/utm-module.type';

@Component({
  selector: 'app-module-integration',
  templateUrl: './module-integration.component.html',
  styleUrls: ['./module-integration.component.css']
})
export class ModuleIntegrationComponent implements OnInit {
  @Input() module: UtmModuleType;
  @Input() serverId: number;
  moduleEnum = UtmModulesEnum;
  currentVersion: string;

  constructor(private checkForUpdatesService: CheckForUpdatesService) {
  }

  ngOnInit() {
    this.getVersionInfo();
  }

  getVersionInfo() {
    this.checkForUpdatesService.getVersion().subscribe(response => {
      this.currentVersion = response.body.version;
    });
  }

}
