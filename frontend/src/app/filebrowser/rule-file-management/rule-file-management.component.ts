import {Component, OnInit} from '@angular/core';

@Component({
    standalone: false,
    selector: 'app-rule-file-management',
    templateUrl: './rule-file-management.component.html',
    styleUrls: ['./rule-file-management.component.css'],
})
export class RuleFileManagementComponent implements OnInit {
  height = window.innerHeight + 250;
  iframeSrc: string;
  server = window.location.hostname;

  constructor() {
  }

  ngOnInit() {
    this.iframeSrc = 'https://' + this.server + '/srv';
  }

}
