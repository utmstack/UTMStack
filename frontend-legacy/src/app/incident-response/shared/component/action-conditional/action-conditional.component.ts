import {Component, EventEmitter, Input, OnInit, Output, ViewChild} from '@angular/core';
import {NgbDropdown} from '@ng-bootstrap/ng-bootstrap';

export enum ActionConditionalEnum  {
  ALWAYS,
  SUCCESS,
  FAILURE
}

@Component({
  selector: 'app-action-conditional',
  templateUrl: './action-conditional.component.html',
  styleUrls: ['./action-conditional.component.scss']
})
export class ActionConditionalComponent implements OnInit {

  options = [
    { key: ActionConditionalEnum.ALWAYS, value: ';'},
    { key: ActionConditionalEnum.SUCCESS, value: '&&'},
    { key: ActionConditionalEnum.FAILURE, value: '||'},
  ];
  actionConditionalEnum = ActionConditionalEnum;
  @Input() option: { key: ActionConditionalEnum, value: string};
  @Output() optionChange = new EventEmitter<{ key: ActionConditionalEnum, value: string}>();
  @ViewChild('dropTracker') dropTracker: NgbDropdown;

  constructor() { }

  ngOnInit() {}

  select(conditional: ActionConditionalEnum) {
    this.option = this.options.find(option => option.key === conditional);
    this.optionChange.emit(this.option);
    this.dropTracker.close();
  }

  getActionConditionalLabel(){
    const conditional = this.option.key;
    switch (conditional) {
      case ActionConditionalEnum.SUCCESS:
        return 'Run on success';

      case ActionConditionalEnum.FAILURE:
        return 'Run on failure';

      default:
        return 'Run always';
    }
  }

  getActionConditionalIcon(){
    const conditional = this.option.key;
    switch (conditional) {
      case ActionConditionalEnum.SUCCESS:
        return 'icon-checkmark-circle text-success';

      case ActionConditionalEnum.FAILURE:
        return 'icon-blocked text-danger fs-8';

      default:
        return 'icon-loop3 text-primary fs-8';
    }
  }
}
