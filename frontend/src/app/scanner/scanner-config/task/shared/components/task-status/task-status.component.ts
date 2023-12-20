import {Component, Input, OnInit} from '@angular/core';
import {TaskModel} from '../../../../../shared/model/task.model';
import {TaskStatusEnum} from '../../enums/task-status.enum';
import {TaskElementViewResolverService} from '../../services/task-element-view-resolver.service';

@Component({
  selector: 'app-task-status',
  templateUrl: './task-status.component.html',
  styleUrls: ['./task-status.component.css']
})
export class TaskStatusComponent implements OnInit {
  @Input() task: TaskModel;
  taskStatusEnum = TaskStatusEnum;

  constructor(public taskElementViewResolverService: TaskElementViewResolverService) {
  }

  ngOnInit() {
  }

}
