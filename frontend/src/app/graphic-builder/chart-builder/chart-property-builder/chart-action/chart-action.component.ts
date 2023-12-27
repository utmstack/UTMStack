import {Component, EventEmitter, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup} from '@angular/forms';
import {ChartActionType} from '../../../../shared/chart/types/action/chart-action.type';

@Component({
  selector: 'app-chart-action',
  templateUrl: './chart-action.component.html',
  styleUrls: ['./chart-action.component.scss']
})
export class ChartActionComponent implements OnInit {
  @Output() chartAction = new EventEmitter<ChartActionType>();
  navigations = ['Events', 'Alerts', 'Vulnerabilities'];
  formAction: FormGroup;

  constructor(private fb: FormBuilder) {
  }

  ngOnInit() {
    this.initFormAction();
    this.formAction.valueChanges.subscribe(action => {
      this.chartAction.emit(action);
    });
    this.chartAction.emit(this.formAction.value);
  }

  initFormAction() {
    this.formAction = this.fb.group({
      active: [true],
      customUrl: [false],
      navigate: [],
      customNavigate: [],
    });
  }

}
