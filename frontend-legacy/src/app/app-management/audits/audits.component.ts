import {HttpResponse} from '@angular/common/http';
import {Component, OnDestroy, OnInit} from '@angular/core';
import {JhiParseLinks} from 'ng-jhipster';
import {ITEMS_PER_PAGE} from '../../shared/constants/pagination.constants';
import {TimeFilterType} from '../../shared/types/time-filter.type';
import {resolveRangeByTime} from '../../shared/util/resolve-date';
import {UtmAuditsService} from './audits.service';
import {UtmAudit} from './shared/type/audit.model';

@Component({
  selector: 'app-audits',
  templateUrl: './audits.component.html',
  styleUrls: ['./audits.component.scss']
})
export class AuditsComponent implements OnInit, OnDestroy {
  audits: UtmAudit[];
  links: any;
  totalItems: number;
  page = 1;
  itemsPerPage = ITEMS_PER_PAGE;
  eventDate: TimeFilterType = resolveRangeByTime('week');
  loading = true;

  constructor(private auditsService: UtmAuditsService,
              private parseLinks: JhiParseLinks) {
    this.itemsPerPage = ITEMS_PER_PAGE;
  }

  ngOnInit() {
    this.loadAll();
  }


  loadAll() {
    const req = {
      page: this.page - 1,
      size: this.itemsPerPage,
      // sort: this.sortBy,
      fromDate: this.eventDate.timeFrom.split('T')[0],
      toDate: this.eventDate.timeTo.split('T')[0]
    };
    this.auditsService
      .query(req)
      .subscribe(
        (res: HttpResponse<UtmAudit[]>) => this.onSuccess(res.body, res.headers),
        (res: HttpResponse<any>) => this.onError(res.body)
      );
  }

  loadPage(page: number) {
    this.page = page;
    this.loadAll();
  }

  ngOnDestroy(): void {
  }

  onFilterChange($event: TimeFilterType) {
    this.eventDate = $event;
    this.loadAll();
  }

  private onSuccess(data, headers) {
    this.links = this.parseLinks.parse(headers.get('link'));
    this.totalItems = headers.get('X-Total-Count');
    this.audits = data;
    this.loading = false;
  }

  private onError(error) {
    // this.alertService.error(error.error, error.message, null);
  }
}

