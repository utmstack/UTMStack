import {HttpResponse} from '@angular/common/http';
import {Component, OnInit} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../shared/alert/utm-toast.service';
import {
  ModalConfirmationComponent
} from '../../shared/components/utm/util/modal-confirmation/modal-confirmation.component';
import {ITEMS_PER_PAGE} from '../../shared/constants/pagination.constants';
import {SortEvent} from '../../shared/directives/sortable/type/sort-event';
import {IncidentResponseVariableService} from '../shared/services/incident-response-variable.service';
import {IncidentVariableType} from '../shared/type/incident-variable.type';
import {IrVariableCreateComponent} from './ir-variable-create/ir-variable-create.component';

@Component({
  selector: 'app-incident-response-variables',
  templateUrl: './incident-response-variables.component.html',
  styleUrls: ['./incident-response-variables.component.scss']
})
export class IncidentResponseVariablesComponent implements OnInit {
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

  deactivateRuleAction(variable: IncidentVariableType, active: boolean) {
    const deleteModalRef = this.modalService.open(ModalConfirmationComponent, {backdrop: 'static', centered: true});
    deleteModalRef.componentInstance.header = 'Deactivate incident response automation';
    deleteModalRef.componentInstance.message = 'Are you sure that you want to deactivate the variable: \n' + variable.variableName;
    deleteModalRef.componentInstance.confirmBtnText = 'Inactive';
    deleteModalRef.componentInstance.confirmBtnIcon = 'icon-cancel-circle2';
    deleteModalRef.componentInstance.confirmBtnType = 'delete';
    deleteModalRef.componentInstance.textDisplay = 'If you inactive this variable, future alerts' +
      ' will not trigger incident response commands.';
    deleteModalRef.componentInstance.textType = 'warning';
    deleteModalRef.result.then(() => {
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


  searchByRule($event: string | number) {
    this.request['name.contains'] = $event;
    this.getVariables();
  }

  searchVariable($event: string | number) {
    this.request['variableName.contains'] = $event;
    this.getVariables();
  }

  editRule(variable: IncidentVariableType) {
  }

  createVariable() {
    const modal = this.modalService.open(IrVariableCreateComponent, {centered: true});
    modal.componentInstance.actionCreated.subscribe(action => {
      this.getVariables();
    });
  }

  getVariablePlaceHolder(variable: IncidentVariableType) {
    return `$(secrets.${variable.variableName})`;
  }
}
