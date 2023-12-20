import {HttpResponse} from '@angular/common/http';
import {Component, OnDestroy, OnInit} from '@angular/core';
import {ActivatedRoute, Router} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import * as moment from 'moment';
import {JhiParseLinks} from 'ng-jhipster';
import {LocalStorageService} from 'ngx-webstorage';
import {ContactUsComponent} from '../../../../shared/components/contact-us/contact-us.component';
import {DEMO_URL} from '../../../../shared/constants/global.constant';
import {ITEMS_PER_PAGE} from '../../../../shared/constants/pagination.constants';
import {SortEvent} from '../../../../shared/directives/sortable/type/sort-event';
import {ActionInitParamsEnum, ActionInitParamsValueEnum} from '../../../../shared/enums/action-init-params.enum';
import {SortByType} from '../../../../shared/types/sort-by.type';
import {AssetSeverityHelpComponent} from '../../../shared/components/asset-severity-help/asset-severity-help.component';
import {TargetModel} from '../../../shared/model/target.model';
import {TaskModel} from '../../../shared/model/task.model';
import {TaskStatusEnum} from '../shared/enums/task-status.enum';
import {TaskTrendEnum} from '../shared/enums/task-trend.enum';
import {TaskElementViewResolverService} from '../shared/services/task-element-view-resolver.service';
import {TaskService} from '../shared/services/task.service';
import {TaskCreateComponent} from '../task-create/task-create.component';
import {TaskDeleteComponent} from '../task-delete/task-delete.component';

@Component({
  selector: 'app-task-list',
  templateUrl: './task-list.component.html',
  styleUrls: ['./task-list.component.scss']
})
export class TaskListComponent implements OnInit, OnDestroy {
  fields: SortByType[] = [
    {
      fieldName: 'Created',
      field: 'created'
    },
    {
      fieldName: 'Task name',
      field: 'name'
    },
    {
      fieldName: 'Target name',
      field: 'target'
    },
    {
      fieldName: 'Status',
      field: 'status'
    },
    {
      fieldName: 'Total reports',
      field: 'reports'
    },
    {
      fieldName: 'Last Run',
      field: 'lastRun'
    },
    {
      fieldName: 'Severity',
      field: 'severity'
    },
    {
      fieldName: 'Trend',
      field: 'trend'
    },
  ];
  tasks: TaskModel[] = [];
  taskShow: TaskModel;
  viewDetail: boolean;
  loading = true;
  totalItems: any;
  page = 1;
  itemsPerPage = ITEMS_PER_PAGE;
  request: any;
  routeData: any;
  links: any;
  predicate: any;
  previousPage: any;
  reverse: any;
  search: string;
  interval: number;
  taskDetail: string;
  sortBy = 'created,desc';

  constructor(
    private activatedRoute: ActivatedRoute,
    private router: Router,
    private parseLinks: JhiParseLinks,
    private modalService: NgbModal,
    private taskService: TaskService,
    private taskElementViewResolverService: TaskElementViewResolverService,
    private localStorage: LocalStorageService) {
  }

  ngOnInit() {
    this.loadTasks();

    this.activatedRoute.queryParams.subscribe(params => {
      if (params[ActionInitParamsEnum.ON_INIT_ACTION]
        && params[ActionInitParamsEnum.ON_INIT_ACTION] === ActionInitParamsValueEnum.SHOW_CREATE_MODAL) {
        this.newTask();
      }
    });

    this.interval = setInterval(() => this.getTask(), 30000);

  }

  loadTasks() {
    this.request = {
      page: this.page - 1,
      size: this.itemsPerPage,
      sort: this.sortBy,
      targets: true
    };
    this.getTask();
  }

  ngOnDestroy(): void {
    clearInterval(this.interval);
  }

  getTask() {
    this.taskService.query(this.request).subscribe(
      (res: HttpResponse<any>) => this.onSuccess(res.body, res.headers),
      (res: HttpResponse<any>) => this.onError(res.body)
    );
  }

  newTask() {
    const modal = this.modalService.open(TaskCreateComponent, {centered: true});
    modal.componentInstance.taskCreated.subscribe(created => {
      this.getTask();
    });
  }

  onSortBy($event: SortEvent) {
    this.sortBy = $event.column + ',' + $event.direction;
    this.loadTasks();
  }

  loadPage(page: any) {
    if (page !== this.previousPage) {
      this.previousPage = page;
      this.transition();
    }
  }

  transition() {
    this.router.navigate(['/scanner/config/task'], {
      queryParams: {
        page: this.page,
        size: this.itemsPerPage,
        sort: this.sortBy
      }
    });
    this.loadTasks();
  }

  runTask(task: TargetModel) {
    if (!window.location.href.includes(DEMO_URL)) {
      this.taskService.start({taskId: task.id}).subscribe(value => {
        this.loadTasks();
      });
    } else {
      this.modalService.open(ContactUsComponent, {centered: true});
    }
  }

  stopTask(task: any) {
    this.taskService.stop({taskId: task.id}).subscribe(value => {
      this.loadTasks();
    });
  }

  resumeTask(task: any) {
    this.taskService.resume({taskId: task.id}).subscribe(value => {
      this.loadTasks();
    });
  }

  editTask(task: any) {
    const modal = this.modalService.open(TaskCreateComponent, {centered: true});
    modal.componentInstance.task = task;
    modal.componentInstance.taskCreated.subscribe(created => {
      this.getTask();
    });
  }

  deleteTask(task: any) {
    const modal = this.modalService.open(TaskDeleteComponent, {centered: true});
    modal.componentInstance.task = task;
    modal.componentInstance.taskDeleted.subscribe(deleted => {
      this.loadTasks();
    });
  }

  resolveClassStatus(status: TaskStatusEnum): string {
    return this.taskElementViewResolverService.resolveClassStatus(status);
  }

  resolveIconClass(status: TaskStatusEnum): string {
    return this.taskElementViewResolverService.resolveIconClass(status);
  }

  resolveTrendClass(trend: TaskTrendEnum): string {
    return this.taskElementViewResolverService.resolveTrendClass(trend);
  }

  resolveTrendIcon(trend: TaskTrendEnum) {
    return this.taskElementViewResolverService.resolveTrendIcon(trend);
  }

  resolveTrendTooltip(trend: TaskTrendEnum) {
    return this.taskElementViewResolverService.resolveTrendTooltip(trend);
  }

  scheduleNextRun(task: TaskModel) {
    // const d = new Date(task.schedule.nextTime);
    return moment(task.schedule.nextTime).format('dddd, MMMM DD YYYY, h:mm:ss a');
    // return d.getFullYear() + '-' + (d.getMonth() + 1) + '-' + d.getDay() + ' ' + d.getHours() + ':' +
    //   d.getMinutes() + ':' + d.getSeconds();
  }

  toggleShowDetail(task: TaskModel) {
    if (task.id === this.taskDetail) {
      this.taskDetail = '';
    } else {
      this.taskDetail = task.id;
    }
  }

  onTaskFilterChange($event: any) {
    Object.keys($event).forEach(key => {
      if ($event[key] !== '' && $event[key] !== null && $event[key] !== 'All') {
        this.request[key] = $event[key];
      } else {
        this.request[key] = undefined;
      }
    });
    this.getTask();
  }

  viewSeverityHelp() {
    const modal = this.modalService.open(AssetSeverityHelpComponent, {centered: true});
  }

  viewTaskResult(task: TaskModel) {
    this.localStorage.store('TASK_RESULT', task);
    this.router.navigate(['/scanner/assets-discovery/task-result'],
      {queryParams: {reportId: task.lastReport.report.uuid}});
  }

  private onSuccess(data, headers) {
    this.links = this.parseLinks.parse(headers.get('link'));
    this.totalItems = headers.get('X-Total-Count');
    this.tasks = data;
    this.loading = false;
    if (this.taskShow) {
      const indexOfTaskShowing = this.tasks.findIndex(value => value.id === this.taskShow.id);
      this.taskShow = this.tasks[indexOfTaskShowing];
    }
  }

  private onError(error) {
    // this.alertService.error(error.error, error.message, null);
  }
}

