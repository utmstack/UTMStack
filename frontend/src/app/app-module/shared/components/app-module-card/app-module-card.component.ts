import {ChangeDetectionStrategy, Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {UtmModuleType} from '../../type/utm-module.type';

@Component({
  selector: 'app-app-module-card',
  templateUrl: './app-module-card.component.html',
  styleUrls: ['./app-module-card.component.scss'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class AppModuleCardComponent implements OnInit {
  @Input() module: UtmModuleType;
  @Output() showModuleIntegration = new EventEmitter<UtmModuleType>();

  constructor() {
  }

  ngOnInit() {
    console.log(this.module.prettyName, this.module.moduleActive);
  }

  showIntegration() {
    this.showModuleIntegration.emit(this.module);
  }
}
