import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {debounceTime} from 'rxjs/operators';
import {UtmToastService} from '../../../../alert/utm-toast.service';
import {InputClassResolve} from '../../../../util/input-class-resolve';
import {IncidentResponseVariableService} from '../../../../services/incidents/incident-response-variable.service';
import {IncidentVariableType} from '../../../../types/incident/incident-variable.type';


@Component({
  selector: 'app-ir-variable-create',
  templateUrl: './ir-variable-create.component.html',
  styleUrls: ['./ir-variable-create.component.scss']
})
export class IrVariableCreateComponent implements OnInit {
  @Input() incidentVariable: IncidentVariableType;
  incidentVariableForm: FormGroup;
  creating = false;
  @Output() actionCreated = new EventEmitter<IncidentVariableType>();
  exist = true;
  typing = true;

  constructor(private formBuilder: FormBuilder,
              public utmToastService: UtmToastService,
              public activeModal: NgbActiveModal,
              public inputClass: InputClassResolve,
              private incidentResponseVariableService: IncidentResponseVariableService) {
  }

  ngOnInit() {
    this.incidentVariableForm = this.formBuilder.group({
      variableName: ['', [Validators.required, Validators.pattern('^[a-zA-Z0-9]*$')]],
      variableValue: ['', Validators.required],
      variableDescription: [''],
      secret: [false, Validators.required],
      id: [{value: null, disabled: true}]
    });
    if (this.incidentVariable) {
      this.exist = false;
      this.incidentVariableForm.patchValue(this.incidentVariable);
      this.incidentVariableForm.get('variableName').disable();
    } else {
      this.incidentVariableForm.get('variableName').valueChanges.pipe(debounceTime(1000)).subscribe(value => {
        this.searchRule(value);
      });
    }
  }

  createVariable() {
    if (this.incidentVariable) {
      this.editAction();
    } else {
      this.createAction();
    }
  }

  createAction() {
    this.creating = true;
    this.incidentResponseVariableService.create(this.incidentVariableForm.value).subscribe((response) => {
      this.utmToastService.showSuccessBottom('Variable saved successfully');
      this.activeModal.close();
      this.creating = false;
      this.actionCreated.emit(response.body);
    }, error1 => {
      this.creating = false;
      this.utmToastService.showError('Error', 'Error saving variable');
    });
  }

  editAction() {
    this.creating = true;
    this.incidentResponseVariableService.update(this.incidentVariableForm.value).subscribe((response) => {
      this.utmToastService.showSuccessBottom('Variable edited successfully');
      this.activeModal.close();
      this.creating = false;
      this.actionCreated.emit(response.body);
    }, error1 => {
      this.creating = false;
      this.utmToastService.showError('Error', 'Error editing Variable');
    });
  }

  searchRule(variableName: string) {
    this.typing = true;
    this.exist = true;
    setTimeout(() => {
      const req = {
        'variableName.equals': variableName
      };
      this.incidentResponseVariableService.query(req).subscribe(response => {
        this.exist = response.body.length > 0;
        this.typing = false;
      });
    }, 1000);
  }


}
