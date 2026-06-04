import {HttpResponse} from '@angular/common/http';
import {Component, OnInit} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../shared/alert/utm-toast.service';
import {ModalConfirmationComponent} from '../../shared/components/utm/util/modal-confirmation/modal-confirmation.component';
import {ITEMS_PER_PAGE} from '../../shared/constants/pagination.constants';
import {SortEvent} from '../../shared/directives/sortable/type/sort-event';
import {IrActionCreateComponent} from '../shared/component/ir-action-create/ir-action-create.component';
import {IncidentResponseActionService} from '../shared/services/incident-response-action.service';
import {IncidentActionType} from '../shared/type/incident-action.type';

@Component({
  selector: 'app-incident-response-command',
  templateUrl: './incident-response-command.component.html',
  styleUrls: ['./incident-response-command.component.scss']
})
export class IncidentResponseCommandComponent implements OnInit {
  actions: IncidentActionType[] = [];
  loading = true;
  page = 1;
  totalItems: number;
  itemsPerPage = ITEMS_PER_PAGE;
  fields: any;
  searching = false;
  private requestParams: any;
  private sortBy: SortEvent;
  private search: string;

  constructor(private incidentResponseActionService: IncidentResponseActionService,
              public utmToastService: UtmToastService,
              private modalService: NgbModal) {
  }

  ngOnInit() {
    this.requestParams = {
      page: this.page - 1,
      size: this.itemsPerPage,
      sort: this.sortBy,
      'actionCommand.contains': ''
    };
    this.getActionsList();
  }

  getActionsList() {
    this.incidentResponseActionService.query(this.requestParams).subscribe(
      (res: HttpResponse<any>) => this.onSuccess(res.body, res.headers),
      (res: HttpResponse<any>) => this.onError(res.body)
    );
  }

  private onSuccess(data, headers) {
    this.totalItems = headers.get('X-Total-Count');
    this.actions = data;
    this.loading = false;
    this.searching = false;
  }

  private onError(body: any) {
    this.loading = false;
    this.searching = false;
  }

  loadPage($event: number) {
    this.requestParams.page = $event - 1;
    this.getActionsList();
  }

  newCommand() {
    const modal = this.modalService.open(IrActionCreateComponent, {centered: true});
    modal.componentInstance.actionCreated.subscribe(action => {
      this.getActionsList();
    });
  }

  onSortBy($event: SortEvent) {
    this.requestParams.sort = $event.column + ',' + $event.direction;
    this.getActionsList();
  }

  searchAction($event: string) {
    const search = $event === '' ? undefined : $event;
    this.requestParams['actionCommand.contains'] = search;
    this.searching = true;
    this.getActionsList();
  }

  openDeleteConfirmation(action: IncidentActionType) {
    const deleteModalRef = this.modalService.open(ModalConfirmationComponent, {centered: true});
    deleteModalRef.componentInstance.header = 'Confirm delete operation';
    deleteModalRef.componentInstance.message = 'Are you sure that you want to delete the action: ' + action.actionCommand;
    deleteModalRef.componentInstance.confirmBtnText = 'Delete';
    deleteModalRef.componentInstance.confirmBtnIcon = 'icon-database-remove';
    deleteModalRef.componentInstance.confirmBtnType = 'delete';
    deleteModalRef.result.then(() => {
      this.deleteAction(action);
    });
  }

  deleteAction(action: IncidentActionType) {
    this.incidentResponseActionService.delete(action.id).subscribe(deleted => {
      this.utmToastService.showSuccessBottom('Action deleted successfully');
      this.getActionsList();
    }, error1 => {
      this.utmToastService.showError('Error', 'Error deleting action');
    });
  }

  editAction(action: IncidentActionType) {
    const modal = this.modalService.open(IrActionCreateComponent, {centered: true});
    modal.componentInstance.action = action;
    modal.componentInstance.actionCreated.subscribe(actionEdited => {
      this.getActionsList();
    });
  }
}
