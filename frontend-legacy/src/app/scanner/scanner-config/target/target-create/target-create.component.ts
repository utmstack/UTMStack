import {AfterViewChecked, ChangeDetectorRef, Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';
import {NgbActiveModal, NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {ITEMS_PER_PAGE} from '../../../../shared/constants/pagination.constants';
import {InputClassResolve} from '../../../../shared/util/input-class-resolve';
import {CredentialModel} from '../../../shared/model/credential.model';
import {PortModel} from '../../../shared/model/port.model';
import {TargetModel} from '../../../shared/model/target.model';
import {CredentialCreateComponent} from '../../credential/credential-create/credential-create.component';
import {CredentialService} from '../../credential/shared/services/credential.service';
import {PortCreateComponent} from '../../port/port-create/port-create.component';
import {PortService} from '../../port/shared/services/port.service';
import {TargetService} from '../shared/services/target.service';

@Component({
  selector: 'app-target-create',
  templateUrl: './target-create.component.html',
  styleUrls: ['./target-create.component.scss']
})
export class TargetCreateComponent implements OnInit, AfterViewChecked {
  @Input() target: TargetModel;
  @Output() targetCreated = new EventEmitter<string>();
  targets: TargetModel[] = [];
  portLists: PortModel[] = [];
  credentials: CredentialModel[];
  aliveTests = [
    'Scan Config Default',
    'ICMP Ping',
    'TCP-ACK Service Ping',
    'TCP-SYN Service Ping',
    'ARP Ping',
    'ICMP & TCP-ACK Service Ping',
    'ICMP & ARP Ping',
    'TCP-ACK Service & ARP Ping',
    'ICMP, TCP-ACK Service & ARP Ping',
    'Consider Alive'
  ];
  targetForm: FormGroup;
  ITEMS_PER_PAGE = ITEMS_PER_PAGE;
  creating = false;

  constructor(public activeModal: NgbActiveModal,
              private fb: FormBuilder,
              public inputClass: InputClassResolve,
              private modalService: NgbModal,
              private portService: PortService,
              private credentialService: CredentialService,
              private cdref: ChangeDetectorRef,
              private utmToastService: UtmToastService,
              private targetService: TargetService) {
  }

  ngOnInit() {
    this.initTargetForm();
    this.getPortList({
      'name.notContains': 'nmap',
      size: 10000
    }, '', true);
    this.getCredentials({
      size: 10000
    });
    if (this.target) {
      this.setFormTarget();
    }
  }

  cleanPortRelatedToOpenvas(ports: PortModel[]): PortModel[] {
    ports.forEach(value => {
      if (value.name.toLowerCase().includes('openvas')) {
        value.name = 'UTMStack Default';
      }
    });
    return ports.filter(value => !value.name.toLowerCase().includes('nmap'));
  }

  setFormTarget() {
    this.targetForm.get('aliveTest').setValue(this.target.aliveTests);
    this.targetForm.get('comment').setValue(this.target.comment);
    this.targetForm.get('name').setValue(this.target.name);
    this.targetForm.get('excludeHosts').setValue(this.target.excludeHosts);
    this.targetForm.get('hosts').setValue(this.target.hosts);
    this.targetForm.get('reverseLookupOnly').setValue(this.target.reverseLookupOnly === '0');
    this.targetForm.get('reverseLookupUnify').setValue(this.target.reverseLookupUnify === '0');
    this.targetForm.get('portList').setValue(this.target.portList !== null ? this.target.portList.uuid : null);
    this.targetForm.get('sshCredentialId').setValue(this.target.sshCredential !== null ? this.target.sshCredential.uuid : null);
    this.targetForm.get('sshCredentialPort').setValue(this.target.sshCredential !== null ? this.target.sshCredential.port : null);
    this.targetForm.get('smbCredential').setValue(this.target.smbCredential !== null ? this.target.smbCredential.uuid : null);
    this.setDisableFields();
  }

  setDisableFields() {
    this.targetForm.get('excludeHosts').disable();
    this.targetForm.get('hosts').disable();
    this.targetForm.get('reverseLookupOnly').disable();
    this.targetForm.get('reverseLookupUnify').disable();
    this.targetForm.get('portList').disable();
    this.targetForm.get('sshCredentialId').disable();
    this.targetForm.get('sshCredentialPort').disable();
    this.targetForm.get('sshCredentialPort').disable();
    this.targetForm.get('smbCredential').disable();
  }

  ngAfterViewChecked(): void {
    this.cdref.detectChanges();
  }

  getPortList(req, setValue?: string, setDefault?: boolean) {
    this.portService.query(req).subscribe(ports => {
      this.portLists = this.cleanPortRelatedToOpenvas(ports.body);
      if (setValue && !setDefault) {
        this.targetForm.get('portList').setValue(setValue);
      }
      if (setDefault) {
        const index = this.portLists.findIndex(value => value.name === 'All IANA assigned TCP and UDP 2012-02-10');
        this.targetForm.get('portList').setValue(this.portLists[index].uuid);
      }
    });
  }

  getCredentials(req) {
    this.credentialService.query(req).subscribe(res => {
        this.credentials = res.body;
      }
    );
  }


  initTargetForm() {
    this.targetForm = this.fb.group({
      aliveTest: [],
      hosts: ['', Validators.required],
      name: ['', Validators.required],
      comment: [''],
      excludeHosts: [''],
      reverseLookupOnly: [false],
      reverseLookupUnify: [false],
      portList: [''],
      smbCredential: [''],
      esxiCredential: [''],
      snmpCredential: [''],
      sshCredentialId: [''],
      sshCredentialPort: [22]
    });
  }

  resolveRequestParams(): any {
    const req = {
      aliveTest: this.targetForm.get('aliveTest').value,
      comment: this.targetForm.get('comment').value,
      excludeHosts: this.targetForm.get('excludeHosts').value,
      name: this.targetForm.get('name').value,
      hosts: this.targetForm.get('hosts').value,
      reverseLookupOnly: this.targetForm.get('reverseLookupOnly').value ? '1' : '0',
      reverseLookupUnify: this.targetForm.get('reverseLookupUnify').value ? '1' : '0',
      sshCredential: null,
      smbCredential: null,
      portList: null,
      targetId: null,
      // esxiCredential: {
      //   id: this.targetForm.get('esxiCredential').value
      // },
      // snmpCredential: {
      //   id: this.targetForm.get('snmpCredential').value,
      // },
    };
    if (this.targetForm.get('portList').value !== '') {
      req.portList = {
        id: this.targetForm.get('portList').value,
      };
    }
    if (this.targetForm.get('sshCredentialId').value !== '') {
      req.sshCredential = {
        id: this.targetForm.get('sshCredentialId').value,
        port: this.targetForm.get('sshCredentialPort').value,
      };
    }
    if (this.targetForm.get('smbCredential').value !== '') {
      req.smbCredential = {
        id: this.targetForm.get('smbCredential').value,
      };
    }
    if (this.target) {
      req.targetId = this.target.uuid;
    }
    return req;
  }

  createTarget() {
    this.creating = true;
    this.targetService.create(this.resolveRequestParams()).subscribe(target => {
      this.utmToastService.showSuccessBottom('Target created successfully');
      this.activeModal.close();
      this.targetCreated.emit(target.body.id);
    }, error1 => {
      this.creating = false;
      this.utmToastService.showError('Error creating target',
        error1.error.statusText);
    });
  }


  editTarget() {
    this.creating = true;
    const req = {
      aliveTest: this.targetForm.get('aliveTest').value,
      comment: this.targetForm.get('comment').value,
      name: this.targetForm.get('name').value,
      targetId: this.target.uuid
    };
    this.targetService.update(req).subscribe(target => {
      this.utmToastService.showSuccessBottom('Target edited successfully');
      this.activeModal.close();
      this.targetCreated.emit(target.body.id);
    }, error1 => {
      this.creating = false;
      this.utmToastService.showError('Error editing target',
        error1.error.statusText);
    });
  }

  newPort() {
    const modal = this.modalService.open(PortCreateComponent, {centered: true});
    modal.componentInstance.portCreated.subscribe(created => {
      this.getPortList({
        size: 10000,
        'name.contains': 'nmap'
      }, created);
    });
  }

  newCredential(type: string) {
    const modal = this.modalService.open(CredentialCreateComponent, {centered: true, backdrop: false});
    modal.componentInstance.type = type;
    modal.componentInstance.credentialCreated.subscribe(created => {
      switch (type) {
        case 'SSH':
          this.targetForm.get('sshCredentialId').setValue(created);
          break;
        case 'SMB':
          this.targetForm.get('smbCredential').setValue(created);
          break;
        case 'ESxi':
          this.targetForm.get('esxiCredential').setValue(created);
          break;
        case 'SNMP':
          this.targetForm.get('snmpCredential').setValue(created);
          break;
      }
      this.getCredentials({
        size: 10000
      });
    });
  }

  getMorePorts() {
  }

  searchPort($event: { term: string; items: any[] }) {
    const req = {
      size: ITEMS_PER_PAGE,
      'name.contains': $event.term
    };
    this.getPortList(req);
  }

  searchCredentials($event: { term: string; items: any[] }) {
  }

  getMoreCredentials() {
  }
}
