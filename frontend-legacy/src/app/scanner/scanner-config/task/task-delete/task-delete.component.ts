import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {TaskModel} from '../../../shared/model/task.model';
import {TaskService} from '../shared/services/task.service';

@Component({
  selector: 'app-task-delete',
  templateUrl: './task-delete.component.html',
  styleUrls: ['./task-delete.component.scss']
})
export class TaskDeleteComponent implements OnInit {
  @Input() task: TaskModel;
  @Output() taskDeleted = new EventEmitter<string>();

  constructor(
    public activeModal: NgbActiveModal,
    private taskService: TaskService,
    private utmToastService: UtmToastService) {
  }

  ngOnInit() {
  }

  deleteTask() {
    this.taskService.delete(this.task.id)
      .subscribe(() => {
        this.utmToastService.showSuccessBottom('Task deleted successfully');
        this.activeModal.close();
        this.taskDeleted.emit('deleted');
      }, (error) => {
        this.utmToastService.showError('Error deleting task',
          error.error.statusText);
      });
  }
}
