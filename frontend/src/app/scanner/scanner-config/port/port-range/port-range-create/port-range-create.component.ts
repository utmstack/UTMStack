import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../../shared/alert/utm-toast.service';
import {InputClassResolve} from '../../../../../shared/util/input-class-resolve';
import {PortModel} from '../../../../shared/model/port.model';
import {PortRangeService} from '../shared/services/port-range.service';

@Component({
  selector: 'app-port-range-edit',
  templateUrl: './port-range-create.component.html',
  styleUrls: ['./port-range-create.component.scss']
})
export class PortRangeCreateComponent implements OnInit {
  @Input() port: PortModel;
  @Output() portRangeAdded = new EventEmitter<string>();
  portRangeForm: FormGroup;
  creating = false;

  constructor(public activeModal: NgbActiveModal,
              private fb: FormBuilder,
              public inputClass: InputClassResolve,
              private portRangeService: PortRangeService,
              private utmToastService: UtmToastService) {
  }

  ngOnInit() {
    this.initFormRangePort();
  }

  initFormRangePort() {
    this.portRangeForm = this.fb.group({
      start: ['', [Validators.required, Validators.min(0)]],
      end: ['', [Validators.required, Validators.min(0)]],
      type: ['', Validators.required],
    });
  }

  createPortRange() {
    const req = {
      end: this.portRangeForm.get('end').value,
      portList: {
        id: this.port.uuid
      },
      start: this.portRangeForm.get('start').value,
      type: this.portRangeForm.get('type').value
    };
    this.creating = true;
    this.portRangeService.create(req).subscribe(portCreated => {
      this.portRangeAdded.emit('success');
      this.activeModal.dismiss();
      this.utmToastService.showSuccessBottom('Port range created successfully');
    }, error1 => {
      this.creating = false;
      this.utmToastService.showError('Error creating port range',
        error1.error.statusText);
    });
  }

  delete() {

  }
}
