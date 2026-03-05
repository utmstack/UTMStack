import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {WinlogbeatService} from '../../../../active-directory/shared/services/winlogbeat.service';
import {ActiveDirectoryTreeType} from '../../../../active-directory/shared/types/active-directory-tree.type';
import {ITEMS_PER_PAGE} from '../../../../shared/constants/pagination.constants';
import {UtmDateFormatEnum} from '../../../../shared/enums/utm-date-format.enum';
import {TimeFilterType} from '../../../../shared/types/time-filter.type';
import {ComplianceControlEvaluationsType} from '../../type/compliance-control-evaluations.type';


@Component({
  selector: 'app-compliance-timeline',
  templateUrl: './compliance-timeline.component.html',
  styleUrls: ['./compliance-timeline.component.scss']
})
export class ComplianceTimelineComponent implements OnInit {
  @Input() evaluations: ComplianceControlEvaluationsType[];
  @Input() time: TimeFilterType;
  @Output() evaluationSelected = new EventEmitter<ComplianceControlEvaluationsType>();
  objectId: ActiveDirectoryTreeType;
  loadingMore = false;
  totalItems: any;
  page = 1;
  itemsPerPage = ITEMS_PER_PAGE;
  filterTime: TimeFilterType;
  loading = true;
  selected: ComplianceControlEvaluationsType;
  formatDateEnum = UtmDateFormatEnum;
  noMoreResult = false;

  constructor(private winlogbeatService: WinlogbeatService) {
  }

  ngOnInit(): void {
    this.selected = null;
    this.page = 1;
    this.totalItems = this.evaluations.length;
    this.loading = false;
  }

  getEvents() {
    if (this.filterTime && this.objectId.objectSid) {
      const req = {
        page: this.page,
        size: this.itemsPerPage,
        sort: '@timestamp,desc',
        sid: this.objectId.objectSid,
        indexPattern: this.objectId.indexPattern,
        from: this.filterTime.timeFrom,
        to: this.filterTime.timeTo,
        'eventId.in': this.evaluations ? this.evaluations.toString() : undefined
      };
      this.winlogbeatService.query(req).subscribe(response => {
        this.loadingMore = false;
        this.loading = false;
        if (response.body === null || response.body.length === 0) {
          this.evaluationSelected.emit(null);
        } else {
          this.evaluations = response.body;
          this.totalItems = Number(response.headers.get('X-Total-Count'));
        }
      });
    } else {
      this.loading = false;
    }
  }

  onScroll() {
    this.loadingMore = true;
    this.page += 1;
    this.getEvents();
  }

  selectEvaluation(evaluation: ComplianceControlEvaluationsType) {
    this.selected = evaluation;
    this.evaluationSelected.emit(evaluation);
  }
}


