import {Component, OnDestroy, OnInit} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {last} from 'rxjs/operators';
import {WorkflowActionsService} from '../../services/workflow-actions.service';
import {ActionTerminalComponent} from '../action-terminal/action-terminal.component';
import {ActionSidebarService} from './action-sidebar.service';

@Component({
  selector: 'app-action-sidebar',
  templateUrl: './action-sidebar.component.html',
  styleUrls: ['./action-sidebar.component.scss']
})
export class ActionSidebarComponent implements OnInit, OnDestroy {

  constructor(private workFlowActionService: WorkflowActionsService,
              public actionSidebarService: ActionSidebarService,
              private modalService: NgbModal) { }

  request = {
    page: 0,
    size: 10,
    'systemOwner.equals': true
  };

  searching: any;
  readonly last = last;

  ngOnInit() {
    this.actionSidebarService.loadData({
      ...this.request,
    });
  }

  addToWorkFlow(action: any) {
    const actionToAdd = {
      ...action,
      id: null
    };
    this.workFlowActionService.addActions(actionToAdd);
  }

  searchReport($event: string ) {
    this.actionSidebarService.loadData({
      ...this.request,
      'label.contains': $event.toString().toString()
    });
  }

  openActionSidebar() {
    const dialogRef = this.modalService.open(ActionTerminalComponent, {centered: true, size: 'lg'});

    dialogRef.result.then(
      result => {
        if (result) {
          this.workFlowActionService.addActions({
            ...result
          });
        }
      },
    );
  }

  onScroll() {

    this.request = {
      ...this.request,
      size: this.request.size + 10
    };

    this.actionSidebarService.loadData({
      ...this.request
    });
  }

  trackByFn(index: number, item: any) {
    return item.id || index;
  }

  ngOnDestroy() {
    this.actionSidebarService.reset();
  }
}
