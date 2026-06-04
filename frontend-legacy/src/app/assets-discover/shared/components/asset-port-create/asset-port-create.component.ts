import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormArray, FormBuilder, FormGroup, Validators} from '@angular/forms';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {NetScanPortsType} from '../../types/net-scan-ports.type';

@Component({
  selector: 'app-asset-port-create',
  templateUrl: './asset-port-create.component.html',
  styleUrls: ['./asset-port-create.component.css']
})
export class AssetPortCreateComponent implements OnInit {
  formPort: FormGroup;
  @Input() ports: NetScanPortsType[];
  @Output() portCreated = new EventEmitter<NetScanPortsType[]>();
  @Output() portValid = new EventEmitter<boolean>();
  portStatus = ['Open', 'Close', 'Filtered', 'Unknown'];

  constructor(private toastService: UtmToastService,
              private fb: FormBuilder) {
  }

  get portFormArray() {
    return this.formPort.controls.portFormArray as FormArray;
  }

  ngOnInit() {
    this.initForm();

    if (this.ports) {
      this.setFormValue();
    }

    this.portFormArray.valueChanges.subscribe(value => {
      if (this.portFormArray.valid) {
        this.portCreated.emit(this.portFormArray.value);
      }
      this.portValid.emit(this.portFormArray.valid);

    });
  }

  setFormValue() {
    for (const port of this.ports) {
      this.addPort(port);
    }
  }

  initForm() {
    this.formPort = this.fb.group({
      portFormArray: this.fb.array([])
    });
  }

  addPort(port?: NetScanPortsType) {
    this.portFormArray.push(this.fb.group({
      port: [port ? port.port : '', [Validators.required]],
      tcp: [port ? port.tcp : '', [Validators.required]],
      udp: [port ? port.udp : '', [Validators.required]],
    }));
  }

  deletePort(index) {
    this.portFormArray.removeAt(index);
    if (!isNaN(index - 1)) {
      this.portFormArray.at(index - 1).get('ports').setValue(null);
    }
  }
}
