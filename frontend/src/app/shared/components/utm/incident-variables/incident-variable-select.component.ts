import {Component, EventEmitter, OnInit, Output} from '@angular/core';
import {IncidentVariableType} from "../../../types/incident/incident-variable.type";
import {IrVariableCreateComponent} from "./ir-variable-create/ir-variable-create.component";
import {IncidentResponseVariableService} from "../../../services/incidents/incident-response-variable.service";
import {NgbModal} from "@ng-bootstrap/ng-bootstrap";
import {ADMIN_ROLE} from "../../../constants/global.constant";

@Component({
  selector: 'app-incident-variable-select',
  template: `
    <div class="w-100">
      <h6 class="font-weight-semibold mb-2">Variables</h6>
      <ul class="w-100" *ngIf="variables && variables.length>0">
        <li *ngFor="let variable of variables"
            (click)="selectVariable(variable);"
            class="cursor-pointer font-size-base d-flex justify-content-between align-items-center mb-1">
                  <span
                    [ngbTooltip]="variable.variableDescription" tooltipClass="utm-tooltip-top">
                      <i [ngClass]="variable.secret === true?'icon-lock2':'icon-cog7'"
                        class="mr-1 font-size-sm"></i>{{ variable.variableName }}
                    </span>
          <code>{{ '$[' + getVariablePlaceHolder(variable) + ']' }}</code>
        </li>
      </ul>
      <div class="d-flex justify-content-center my-3" *appHasAnyAuthority="roles">
        <span class="text-blue-800 cursor-pointer" (click)="createVariable()"><i class="icon-plus2 mr-1"></i> Create variable</span>
      </div>
    </div>
  `
})

export class IncidentVariableSelectComponent implements OnInit {
  @Output() variableSelected = new EventEmitter<string>();
  variables: IncidentVariableType[];
  roles = [ADMIN_ROLE];

  constructor(private incidentResponseVariableService: IncidentResponseVariableService,
              private modalService: NgbModal,) {
  }

  ngOnInit() {
    this.getVariables();
  }

  getVariables() {
    this.incidentResponseVariableService.query({page: 0, size: 100}).subscribe(response => {
      if (response.body) {
        this.variables = response.body;
      }
    });
  }

  getVariablePlaceHolder(variable: IncidentVariableType) {
    return `variables.${variable.variableName}`;
  }

  createVariable() {
    const modal = this.modalService.open(IrVariableCreateComponent, {centered: true});
    modal.componentInstance.actionCreated.subscribe(() => {
      this.getVariables();
    });
  }

  selectVariable(variable: IncidentVariableType) {
    this.variableSelected.emit(`variables.${variable.variableName}`)
  }
}
