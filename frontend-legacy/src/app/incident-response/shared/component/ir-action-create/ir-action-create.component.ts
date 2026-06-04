import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {InputClassResolve} from '../../../../shared/util/input-class-resolve';
import {IncidentResponseActionService} from '../../services/incident-response-action.service';
import {IncidentActionType} from '../../type/incident-action.type';

@Component({
  selector: 'app-incident-response-command-create',
  templateUrl: './ir-action-create.component.html',
  styleUrls: ['./ir-action-create.component.scss']
})
export class IrActionCreateComponent implements OnInit {
  @Input() action: IncidentActionType;
  @Output() actionCreated = new EventEmitter<IncidentActionType>();
  creating: any;
  formCommand: FormGroup;

  constructor(public activeModal: NgbActiveModal,
              public inputClass: InputClassResolve,
              public utmToastService: UtmToastService,
              private incidentResponseActionService: IncidentResponseActionService,
              private fb: FormBuilder) {
  }

  ngOnInit() {
    this.initFormCommand();
    if (this.action) {
      this.formCommand.patchValue(this.action);
    }
  }

  initFormCommand() {
    this.formCommand = this.fb.group({
      actionCommand: ['', [Validators.required]],
      actionDescription: ['', [Validators.required]],
      actionParams: [],
      actionType: [8, [Validators.required]],
      actionEditable: [true, [Validators.required]],
      id: []
    });
  }

  createCommand() {
    if (this.action) {
      this.editAction();
    } else {
      this.createAction();
    }
  }

  createAction() {
    this.creating = true;
    this.incidentResponseActionService.create(this.formCommand.value).subscribe((response) => {
      this.utmToastService.showSuccessBottom('Action saved successfully');
      this.activeModal.close();
      this.creating = false;
      this.actionCreated.emit(response.body);
    }, error1 => {
      this.creating = false;
      this.utmToastService.showError('Error', 'Error saving action');
    });
  }

  editAction() {
    this.creating = true;
    this.incidentResponseActionService.update(this.formCommand.value).subscribe((response) => {
      this.utmToastService.showSuccessBottom('Action edited successfully');
      this.activeModal.close();
      this.creating = false;
      this.actionCreated.emit(response.body);
    }, error1 => {
      this.creating = false;
      this.utmToastService.showError('Error', 'Error editing action');
    });
  }
}
