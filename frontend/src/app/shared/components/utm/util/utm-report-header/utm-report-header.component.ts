import {Component, Input, OnInit} from '@angular/core';
import {DomSanitizer} from '@angular/platform-browser';
import * as moment from 'moment';
import {ThemeChangeBehavior} from '../../../../behaviors/theme-change.behavior';

@Component({
  selector: 'app-utm-report-header',
  templateUrl: './utm-report-header.component.html',
  styleUrls: ['./utm-report-header.component.scss']
})
export class UtmReportHeaderComponent implements OnInit {
  @Input() reportName: string;
  date: any;
  @Input() subheader?: string;
  reportIcon: string;

  constructor(private themeChangeBehavior: ThemeChangeBehavior,
              public sanitizer: DomSanitizer) {
    this.themeChangeBehavior.$themeReportIcon.subscribe(icon => {
      this.reportIcon = icon;
    });
  }

  ngOnInit() {
    this.date = moment(new Date()).format('YYYY-MM-DD HH:MM:SS');
  }

}
