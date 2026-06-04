import {Injectable} from '@angular/core';
import {TaskStatusEnum} from '../enums/task-status.enum';
import {TaskTrendEnum} from '../enums/task-trend.enum';

@Injectable({
  providedIn: 'root'
})
export class TaskElementViewResolverService {

  constructor() {
  }

  resolveClassStatus(status: TaskStatusEnum): string {
    // Delete Requested|Done|New|Requested|Running|Stop Requested|Stopped|Internal Error
    switch (status) {
      case TaskStatusEnum.DONE:
        return 'badge-flat border-primary-800 text-primary-800';
      case TaskStatusEnum.NEW:
        return 'badge-flat border-success-800 text-success-800';
      case TaskStatusEnum.DELETE_REQUESTED:
        return 'badge-flat border-danger-800 text-danger-800';
      case TaskStatusEnum.REQUESTED:
        return 'badge-flat border-slate-800 text-slate-800';
      case TaskStatusEnum.STOP_REQUESTED:
        return 'badge-flat border-grey-800 text-grey-800';
      case TaskStatusEnum.STOPPED:
        return 'badge-flat border-warning-800 text-warning-800';
      default:
        return 'badge-flat border-danger-800 text-danger-800';
    }
  }

  resolveIconClass(status: TaskStatusEnum): string {
    switch (status) {
      case TaskStatusEnum.DONE:
        return 'icon-shield-check';
      case TaskStatusEnum.NEW:
        return 'icon-alarm-check';
      case TaskStatusEnum.DELETE_REQUESTED:
        return 'icon-close2';
      case TaskStatusEnum.REQUESTED:
        return 'icon-drive';
      case TaskStatusEnum.STOP_REQUESTED:
        return 'icon-stop';
      case TaskStatusEnum.STOPPED:
        return 'icon-stop2';
      default:
        return 'icon-warning22';
    }
  }

  // up|down|more|less|same

  resolveTrendClass(trend: TaskTrendEnum): string {
    if (trend === TaskTrendEnum.UP || trend === TaskTrendEnum.MORE) {
      return 'badge-danger';
    } else if (trend === TaskTrendEnum.LESS || trend === TaskTrendEnum.DOWN) {
      return 'badge-success';
    } else {
      return 'badge-primary';
    }
  }

  resolveTrendIcon(trend: TaskTrendEnum) {
    if (trend === TaskTrendEnum.UP || trend === TaskTrendEnum.MORE) {
      return 'icon-stats-growth2';
    } else if (trend === TaskTrendEnum.LESS || trend === TaskTrendEnum.DOWN) {
      return 'icon-stats-decline2';
    } else {
      return 'icon-arrow-right8';
    }
  }

  resolveTrendTooltip(trend: TaskTrendEnum) {
    if (trend === TaskTrendEnum.UP || trend === TaskTrendEnum.MORE) {
      return 'Severity increased';
    } else if (trend === TaskTrendEnum.LESS || trend === TaskTrendEnum.DOWN) {
      return 'Severity decreased';
    } else {
      return 'Vulnerabilities did not change';
    }
  }
}
