import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {UtmModuleType} from '../../type/utm-module.type';

@Component({
  selector: 'app-app-module-card',
  templateUrl: './app-module-card.component.html',
  styleUrls: ['./app-module-card.component.scss']
})
export class AppModuleCardComponent implements OnInit {
  @Input() module: UtmModuleType;
  @Output() showModuleIntegration = new EventEmitter<UtmModuleType>();

  constructor() {
  }

  ngOnInit() {
  }

  showIntegration() {
    this.showModuleIntegration.emit(this.module);
  }
}
