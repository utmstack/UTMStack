import {Component, OnDestroy, OnInit} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {WorkflowActionsService} from '../../services/workflow-actions.service';
import {ActionTerminalComponent} from '../action-terminal/action-terminal.component';
import {ActionSidebarService} from './action-sidebar.service';
import {last} from "rxjs/operators";

@Component({
  selector: 'app-action-sidebar',
  templateUrl: './action-sidebar.component.html',
  styleUrls: ['./action-sidebar.component.scss']
})
export class ActionSidebarComponent implements OnInit, OnDestroy {

  request = {
    page: 0,
    size: 10,
    'systemOwner.equals': true
  };

  searching: any;

  constructor(private workFlowActionService: WorkflowActionsService,
              public actionSidebarService: ActionSidebarService,
              private modalService: NgbModal) { }

  ngOnInit() {
    this.actionSidebarService.loadData({
      ...this.request,
    });
  }

  addToWorkFlow(action: any) {
    this.workFlowActionService.addActions(action);
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
    this.actionSidebarService.loadData({
      ...this.request,
      size: this.request.size + 10,
    });
  }

  trackByFn(index: number, item: any) {
    return item.id || index;
  }

  ngOnDestroy() {
    this.actionSidebarService.reset();
  }

  protected readonly last = last;
}
