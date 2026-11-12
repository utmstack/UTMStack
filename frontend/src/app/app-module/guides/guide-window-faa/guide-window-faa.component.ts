import {Component, Input, OnInit} from '@angular/core';

@Component({
    standalone: false,
    selector: 'app-guide-window-faa',
    templateUrl: './guide-window-faa.component.html',
    styleUrls: ['./guide-window-faa.component.scss'],
})
export class GuideWindowFaaComponent implements OnInit {
  @Input() integrationId: number;
  @Input() serverId: number;

  constructor() {
  }

  ngOnInit() {
  }

}
