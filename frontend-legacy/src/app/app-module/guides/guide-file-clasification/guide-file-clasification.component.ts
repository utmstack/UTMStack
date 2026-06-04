import {Component, Input, OnInit} from '@angular/core';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';

@Component({
  selector: 'app-guide-file-clasification',
  templateUrl: './guide-file-clasification.component.html',
  styleUrls: ['./guide-file-clasification.component.scss']
})
export class GuideFileClasificationComponent implements OnInit {
  @Input() integrationId: number;
  module = UtmModulesEnum;
  @Input() serverId: number;

  constructor() {
  }

  ngOnInit() {
  }

}
