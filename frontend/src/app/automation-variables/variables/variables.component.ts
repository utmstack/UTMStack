import {HttpResponse} from '@angular/common/http';
import {Component, OnInit} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../shared/alert/utm-toast.service';
import {
  ModalConfirmationComponent
} from '../../shared/components/utm/util/modal-confirmation/modal-confirmation.component';
import {ITEMS_PER_PAGE} from '../../shared/constants/pagination.constants';
import {SortEvent} from '../../shared/directives/sortable/type/sort-event';
import {IrVariableCreateComponent} from '../../shared/components/utm/incident-variables/ir-variable-create/ir-variable-create.component';
import {IncidentResponseVariableService} from '../../shared/services/incidents/incident-response-variable.service';
import {IncidentVariableType} from '../../shared/types/incident/incident-variable.type';

@Component({
  selector: 'app-variables',
  templateUrl: './variables.component.html',
  styleUrls: ['./variables.component.scss']
})
export class VariablesComponent implements OnInit {
  loading = true;
  variables: IncidentVariableType[];
  itemsPerPage = 15;
  totalItems: number;
  request = {
    page: 0,
    size: ITEMS_PER_PAGE,
    sort: '',
    'variableName.contains': null,
  };

  constructor(
    private modalService: NgbModal,
    private utmToastService: UtmToastService,
    private incidentResponseVariableService: IncidentResponseVariableService
  ) {
  }

  ngOnInit() {
    this.getVariables();
  }

  deactivateAction(variable: IncidentVariableType) {
    const deleteModalRef = this.modalService.open(ModalConfirmationComponent, {backdrop: 'static', centered: true});
    deleteModalRef.componentInstance.header = 'Delete automation variable';
    deleteModalRef.componentInstance.message = 'Are you sure that you want to delete the variable: \n' + variable.variableName + '?';
    deleteModalRef.componentInstance.confirmBtnText = 'Delete';
    deleteModalRef.componentInstance.confirmBtnIcon = 'icon-cancel-circle2';
    deleteModalRef.componentInstance.confirmBtnType = 'delete';
    deleteModalRef.componentInstance.textDisplay = 'Incident response automation could encounter failures' +
      ' when attempting to reference this specific variable.';
    deleteModalRef.componentInstance.textType = 'warning';
    deleteModalRef.result.then(() => {
      this.delete(variable);
    });
  }

  loadPage(page: number) {
    this.request.page = page - 1;
    this.getVariables();
  }

  onItemsPerPageChange($event: number) {
    this.request.size = $event;
    this.getVariables();
  }

  getVariables() {
    this.incidentResponseVariableService.query(this.request).subscribe(
      (res: HttpResponse<any>) => this.onSuccess(res.body, res.headers),
      (res: HttpResponse<any>) => this.onError(res.body)
    );
  }

  private onSuccess(data, headers) {
    this.totalItems = headers.get('X-Total-Count');
    this.variables = data;
    this.loading = false;
  }

  private onError(error) {
    // this.alertService.error(error.error, error.message, null);
  }

  onSort($event: SortEvent) {
    this.request.sort = $event.column + ',' + $event.direction;
    this.getVariables();
  }

  searchVariable($event: string | number) {
    this.request['variableName.contains'] = $event;
    this.getVariables();
  }

  createVariable() {
    const modal = this.modalService.open(IrVariableCreateComponent, {centered: true});
    modal.componentInstance.actionCreated.subscribe(action => {
      this.getVariables();
    });
  }

  editVariable(variable: IncidentVariableType) {
    const modal = this.modalService.open(IrVariableCreateComponent, {centered: true});
    modal.componentInstance.incidentVariable = variable;
    modal.componentInstance.actionCreated.subscribe(action => {
      this.getVariables();
    });
  }

  getVariablePlaceHolder(variable: IncidentVariableType) {
    return `$[variables.${variable.variableName}]`;
  }

  delete(variable: IncidentVariableType) {
    this.incidentResponseVariableService.delete(variable.id).subscribe(() => {
      this.utmToastService.showSuccessBottom('Variable deleted successfully');
      this.getVariables();
    }, error1 => {
      this.utmToastService.showError('Error', 'Error deleting variable');
    });
  }
}
