import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../shared/alert/utm-toast.service';
import {LogAnalyzerQueryService} from '../../shared/services/log-analyzer-query.service';
import {LogAnalyzerQueryType} from '../../shared/type/log-analyzer-query.type';

@Component({
  selector: 'app-log-analyzer-query-delete',
  templateUrl: './log-analyzer-query-delete.component.html',
  styleUrls: ['./log-analyzer-query-delete.component.scss']
})
export class LogAnalyzerQueryDeleteComponent implements OnInit {
  @Input() query: LogAnalyzerQueryType;
  @Output() queryDeleted = new EventEmitter<string>();

  constructor(public activeModal: NgbActiveModal,
              private utmToastService: UtmToastService,
              private analyzerQueryService: LogAnalyzerQueryService) {
  }

  ngOnInit() {
  }

  deleteQuery() {
    this.analyzerQueryService.delete(this.query.id)
      .subscribe(() => {
        this.utmToastService.showSuccessBottom('Query deleted successfully');
        this.activeModal.close();
        this.queryDeleted.emit('deleted');
      }, (error) => {
        this.utmToastService.showError('Error deleting query',
          error.error.statusText);
      });
  }

}
