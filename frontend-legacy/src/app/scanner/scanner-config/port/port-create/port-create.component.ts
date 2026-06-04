import {Component, EventEmitter, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {InputClassResolve} from '../../../../shared/util/input-class-resolve';
import {portValidator} from '../shared/form-validator/port-form.validator';
import {PortService} from '../shared/services/port.service';


@Component({
  selector: 'app-port-create',
  templateUrl: './port-create.component.html',
  styleUrls: ['./port-create.component.scss']
})
export class PortCreateComponent implements OnInit {
  @Output() portCreated = new EventEmitter<string>();
  edit = false;
  formPortList: FormGroup;
  creating = false;

  constructor(public activeModal: NgbActiveModal,
              private fb: FormBuilder,
              public inputClass: InputClassResolve,
              private portService: PortService,
              private utmToastService: UtmToastService) {
  }

  ngOnInit() {
    this.initFormPort();
  }

  initFormPort() {
    this.formPortList = this.fb.group({
      name: ['', [Validators.required]],
      comment: ['', Validators.required],
      tcpSingle: ['', [Validators.required, portValidator()]],
      udpSingle: ['', [Validators.required, portValidator()]],
    });
  }

  createPort() {
    this.creating = true;
    const port = {
      comment: this.formPortList.get('comment').value,
      name: this.formPortList.get('name').value,
      portRange: 'T:' + this.formPortList.get('tcpSingle').value +
        ',U:' + this.formPortList.get('udpSingle').value
    };
    this.portService.create(port).subscribe(portCreated => {
      this.portCreated.emit(portCreated.body.id);
      this.activeModal.dismiss();
      this.utmToastService.showSuccessBottom('Port list created successfully');
    }, error1 => {
      this.creating = false;
      this.utmToastService.showError('Error creating port list',
        'Error creating port list, check your network');
    });
  }

  editPort() {
  }

  newPort(ssh: string) {
  }
}
