import {HttpResponse} from '@angular/common/http';
import {Component, OnInit} from '@angular/core';
import {ActivatedRoute, Router} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {JhiParseLinks} from 'ng-jhipster';
import {ITEMS_PER_PAGE} from '../../../../shared/constants/pagination.constants';
import {SortEvent} from '../../../../shared/directives/sortable/type/sort-event';
import {SortByType} from '../../../../shared/types/sort-by.type';
import {UsedByComponent} from '../../../shared/components/used-by/used-by.component';
import {ScheduleModel} from '../../../shared/model/schedule.model';
import {ScheduleCreateComponent} from '../schedule-create/schedule-create.component';
import {ScheduleDeleteComponent} from '../schedule-delete/schedule-delete.component';
import {ScheduleService} from '../shared/services/schedule.service';

@Component({
  selector: 'app-schedule-list',
  templateUrl: './schedule-list.component.html',
  styleUrls: ['./schedule-list.component.scss']
})
export class ScheduleListComponent implements OnInit {
  fields: SortByType[] = [
    {
      fieldName: 'Name',
      field: 'name'
    },
    {
      fieldName: 'First run',
      field: 'type'
    },
    {
      fieldName: 'Next Run',
      field: 'insecureUse'
    },
    {
      fieldName: 'Period',
      field: 'login'
    },
    {
      fieldName: 'Duration',
      field: 'login'
    }
  ];
  schedules: ScheduleModel[] = [];
  loading = false;
  totalItems: any;
  page: any;
  itemsPerPage = ITEMS_PER_PAGE;
  error: any;
  success: any;
  routeData: any;
  links: any;
  predicate: any;
  previousPage: any;
  reverse: any;
  private sortBy: string;
  private requestParams: any;

  constructor(private modalService: NgbModal,
              private activatedRoute: ActivatedRoute,
              private router: Router,
              private parseLinks: JhiParseLinks,
              private scheduleService: ScheduleService) {
  }

  ngOnInit() {
    this.routeData = this.activatedRoute.data.subscribe(data => {
      this.page = data.pagingParams.page;
      this.previousPage = data.pagingParams.page;
      this.reverse = data.pagingParams.ascending;
      this.predicate = data.pagingParams.predicate;
      // this.sortBy = this.predicate + ',' + (this.reverse ? 'asc' : 'desc');
    });
    this.requestParams = {
      page: this.page - 1,
      size: this.itemsPerPage,
      sort: this.sortBy,
      tasks: true
    };
    this.getSchedules();
  }

  newSchedule() {
    const modal = this.modalService.open(ScheduleCreateComponent, {centered: true});
    modal.componentInstance.scheduleCreated.subscribe(created => {
      this.getSchedules();
    });
  }

  editSchedule(schedule?: any) {
    const modal = this.modalService.open(ScheduleCreateComponent, {centered: true});
    modal.componentInstance.schedule = schedule;
    modal.componentInstance.scheduleCreated.subscribe(created => {
      this.getSchedules();
    });
  }

  deleteSchedule(schedule: any) {
    const modal = this.modalService.open(ScheduleDeleteComponent, {centered: true});
    modal.componentInstance.schedule = schedule;
    modal.componentInstance.scheduleDeleted.subscribe(deleted => {
      this.getSchedules();
    });
  }

  onSortBy($event: SortEvent) {
    this.sortBy = $event.column + ',' + $event.direction;
    this.getSchedules();
  }


  getSchedules() {
    this.loading = true;
    this.scheduleService.query(this.requestParams).subscribe(
      (res: HttpResponse<any>) => this.onSuccess(res.body, res.headers),
      (res: HttpResponse<any>) => this.onError(res.body)
    );
  }

  loadPage(page: number) {
    if (page !== this.previousPage) {
      this.previousPage = page;
      this.transition();
    }
  }

  transition() {
    this.router.navigate(['/scanner/config//schedule'], {
      queryParams: {
        page: this.page,
        sort: this.sortBy
      }
    });
    this.getSchedules();
  }

  resolveDuration(duration): string {
    const time = duration / 60 / 60;
    if (time <= 24) {
      return time + ' hour(s)';
    } else if (time <= 1) {
      return time + ' days(s)';
    }
  }

  showUse(schedule: ScheduleModel) {
    const modal = this.modalService.open(UsedByComponent, {centered: true});
    modal.componentInstance.using = schedule.tasks;
    modal.componentInstance.dependency = 'Tasks';
    modal.componentInstance.type = 'schedule';
    modal.componentInstance.name = schedule.name;
  }

  onScheduleFilterChange($event: any) {
    Object.keys($event).forEach(key => {
      if ($event[key] !== '' && $event[key] !== null) {
        this.requestParams[key] = $event[key];
      } else {
        this.requestParams[key] = undefined;
      }
    });
    this.getSchedules();
  }

  private onSuccess(data, headers) {
    this.links = this.parseLinks.parse(headers.get('link'));
    this.totalItems = headers.get('X-Total-Count');
    this.schedules = data;
    this.loading = false;
  }

  private onError(error) {
    // this.alertService.error(error.error, error.message, null);
  }
}
