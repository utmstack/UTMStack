import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {UtmToastService} from '../../../shared/alert/utm-toast.service';
import {LogstashService} from '../../../shared/services/logstash/logstash.service';
import {LogstashFilterType} from '../../shared/types/logstash-filter.type';
import {UtmPipeline} from '../../shared/types/logstash-stats.type';

@Component({
  selector: 'app-logstash-filter-create',
  templateUrl: './logstash-filter-create.component.html',
  styleUrls: ['./logstash-filter-create.component.scss']
})
export class LogstashFilterCreateComponent implements OnInit {
  @Output() filterEditClose = new EventEmitter<LogstashFilterType>();
  @Output() close = new EventEmitter<string>();
  @Input() filter: LogstashFilterType;
  @Input() pipeline: UtmPipeline;
  saving = false;
  show = true;
  error = false;

  constructor(private logstashFilterService: LogstashService,
              private utmToastService: UtmToastService) {
  }

  ngOnInit() {
    if (!this.filter || !this.filter.id) {
      this.filter = {logstashFilter: '', filterName: ''};
    }

  }


  save() {
    this.saving = true;
    if (this.filter.id) {
      this.updateFilter();
    } else {
      this.createFilter();
    }
  }

  updateFilter() {
    this.logstashFilterService.update(this.filter).subscribe(response => {
      this.onSaveSuccess('Log filter updated successfully', response.body);
    }, error => {
      this.onSaveError();
    });
  }

  onSaveSuccess(text: string, response: any) {
    this.utmToastService.showSuccessBottom(text);
    this.saving = false;
    this.filterEditClose.emit(response);
  }

  onSaveError() {
    this.saving = false;
    this.utmToastService.showError('Error saving log filter', 'Error while trying to change log filter');
  }

  createFilter() {
    this.logstashFilterService.create(this.filter, {pipelineId: this.pipeline.id}).subscribe(response => {
      this.onSaveSuccess('Log filter created successfully', response.body);
    }, error => {
      this.onSaveError();
    });
  }


}
