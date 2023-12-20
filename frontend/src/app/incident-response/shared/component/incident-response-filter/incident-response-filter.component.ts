import {Component, EventEmitter, OnInit, Output} from '@angular/core';
import {TimeFilterType} from '../../../../shared/types/time-filter.type';
import {IR_STATUS} from '../../const/ir-status.const';
import {IncidentResponseActionService} from '../../services/incident-response-action.service';
import {IncidentActionType} from '../../type/incident-action.type';

@Component({
  selector: 'app-incident-response-filter',
  templateUrl: './incident-response-filter.component.html',
  styleUrls: ['./incident-response-filter.component.scss']
})
export class IncidentResponseFilterComponent implements OnInit {
  @Output() irFilterChange = new EventEmitter<any>();
  defaultTimeEvent: TimeFilterType = {range: 'all', timeFrom: null, timeTo: null};
  filter = {};
  actions: IncidentActionType[];
  status = IR_STATUS;

  constructor(private incidentResponseActionService: IncidentResponseActionService) {
  }

  ngOnInit() {
    this.getActionsList();
  }

  onFilter($event: string) {
    this.filter['agent.contains'] = $event === '' ? undefined : $event;
    this.irFilterChange.emit(this.filter);
  }

  getActionsList() {
    this.incidentResponseActionService.query({page: 0, size: 500}).subscribe(res => {
      this.actions = res.body;
    });
  }

  searchByCreation($event: TimeFilterType) {
    this.filter['created.greaterThan'] = $event.timeFrom.substring(0, $event.timeFrom.lastIndexOf(':'));
    this.filter['created.lessThan'] = $event.timeTo.substring(0, $event.timeFrom.lastIndexOf(':'));
    this.irFilterChange.emit(this.filter);
  }
}
