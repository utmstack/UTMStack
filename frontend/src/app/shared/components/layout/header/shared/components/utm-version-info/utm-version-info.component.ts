import {Component, Input, OnInit} from '@angular/core';
import {VersionInfo} from '../../../../../../types/updates/updates.type';

@Component({
  selector: 'app-utm-version-info',
  templateUrl: './utm-version-info.component.html',
  styleUrls: ['./utm-version-info.component.css']
})
export class UtmVersionInfoComponent implements OnInit {
  @Input('version') currentVersion: VersionInfo;

  ngOnInit() {}

}
