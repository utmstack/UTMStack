import {Component, Input, OnInit} from '@angular/core';
import {TaskModel} from '../../../shared/model/task.model';
import {TaskElementViewResolverService} from '../shared/services/task-element-view-resolver.service';

@Component({
  selector: 'app-task-detail',
  templateUrl: './task-detail.component.html',
  styleUrls: ['./task-detail.component.scss']
})
export class TaskDetailComponent implements OnInit {
  @Input() task: TaskModel;

  constructor(public taskElementViewResolverService: TaskElementViewResolverService) {
  }

  ngOnInit() {
  }

}
