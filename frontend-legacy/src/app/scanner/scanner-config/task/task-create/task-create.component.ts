import {HttpResponse} from '@angular/common/http';
import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';
import {NgbActiveModal, NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {Observable} from 'rxjs';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {InputClassResolve} from '../../../../shared/util/input-class-resolve';
import {ScannerConfigModel} from '../../../shared/model/scanner-config.model';
import {ScannerModel} from '../../../shared/model/scanner.model';
import {ScheduleModel} from '../../../shared/model/schedule.model';
import {TargetModel} from '../../../shared/model/target.model';
import {TaskModel} from '../../../shared/model/task.model';
import {ScheduleCreateComponent} from '../../schedule/schedule-create/schedule-create.component';
import {ScheduleService} from '../../schedule/shared/services/schedule.service';
import {TargetService} from '../../target/shared/services/target.service';
import {TargetCreateComponent} from '../../target/target-create/target-create.component';
import {ConfigScannerService} from '../shared/services/config-scanner.service';
import {ScannerService} from '../shared/services/scanner.service';
import {TaskService} from '../shared/services/task.service';

@Component({
  selector: 'app-task-create',
  templateUrl: './task-create.component.html',
  styleUrls: ['./task-create.component.scss']
})
export class TaskCreateComponent implements OnInit {
  @Output() taskCreated = new EventEmitter<string>();
  @Input() task: TaskModel;
  @Input() mode: 'simple' | 'normal' = 'normal';
  taskForm: FormGroup;
  hostOrder = ['Sequential', 'Reverse', 'Random'];
  targets: TargetModel[] = [];
  schedules: ScheduleModel[] = [];
  scanners: ScannerModel[] = [];
  configs: ScannerConfigModel[] = [];
  loadingTarget = true;
  loadingSchedules = true;
  loadingScanners = true;
  loadingScannersConfig = true;
  runOnce = true;
  creating = false;

  constructor(
    public activeModal: NgbActiveModal,
    private fb: FormBuilder,
    public inputClass: InputClassResolve,
    private modalService: NgbModal,
    private targetService: TargetService,
    private taskService: TaskService,
    private utmToastService: UtmToastService,
    private scheduleService: ScheduleService,
    private scannerService: ScannerService,
    private configScannerService: ConfigScannerService) {
  }

  ngOnInit() {
    this.initTaskForm();
    this.getTargets();
    this.getSchedules();
    this.getScanners().subscribe(() => {
      if (!this.task) {
        this.taskForm.get('scannerId').setValue(this.scanners[1].uuid);
      }
    });
    this.getScannerConfig().subscribe(() => {
      if (!this.task) {
        const indexDefaultConfig = this.configs.findIndex(value => value.name === 'Full and fast');
        this.taskForm.get('configId').setValue(this.configs[indexDefaultConfig].uuid);
      }
    });
    if (this.task) {
      this.setTaskForm();
    }
    if (this.mode === 'simple') {
      this.setSimpleScanForm();
    }
  }


  getTargets() {
    this.targetService.query({size: 1000}).subscribe(
      (res: HttpResponse<any>) => {
        this.targets = res.body;
        this.loadingTarget = false;
      },
      (res: HttpResponse<any>) => {
        this.loadingTarget = false;
      }
    );
  }

  initTaskForm() {
    this.taskForm = this.fb.group({
      id: [],
      name: ['', Validators.required],
      comment: [''],
      targetId: ['', Validators.required],
      scheduleId: [],
      schedulePeriods: [],
      inAssets: [1],
      applyOverrides: [1],
      minQod: [70],
      alterable: [1],
      autoDelete: [0],
      autoDeleteData: [1],
      scannerId: [],
      scannerType: [2],
      configId: [],
      sourceIface: [],
      hostsOrdering: ['Sequential'],
      maxChecks: [4],
      maxHosts: [20]
    });
  }

  extractFromPreferences(property: string) {
    const index = this.task.preferences.findIndex(value => value.scannerName === property);
    if (this.task.preferences[index].value === 'yes' || this.task.preferences[index].value === 'no') {
      return this.task.preferences[index].value === 'yes' ? 1 : 0;
    } else {
      return this.task.preferences[index].value;
    }
  }

  getSchedules() {
    this.scheduleService.query({size: 1000}).subscribe(
      (res: HttpResponse<any>) => {
        this.schedules = res.body;
        this.loadingSchedules = false;
      },
      (res: HttpResponse<any>) => {
        this.loadingSchedules = false;
      }
    );
  }

  getScanners(): Observable<boolean> {
    return new Observable<boolean>(subscriber => {
      this.scannerService.query({size: 1000}).subscribe(
        (res: HttpResponse<any>) => {
          this.scanners = res.body;
          this.scanners = this.cleanDataRelatedToOpenvas(res.body);
          subscriber.next(true);
          this.loadingScanners = false;
        },
        (res: HttpResponse<any>) => {
          this.loadingScanners = false;
          subscriber.next(false);
        }
      );
    });
  }

  cleanDataRelatedToOpenvas(scanner: ScannerModel[]): ScannerModel[] {
    scanner.forEach(value => {
      if (value.name.toLowerCase().includes('openvas')) {
        value.name = 'UTMStack Scanner';
      }
    });
    return scanner.filter(value => !value.name.toLowerCase().includes('nmap'));
  }

  getScannerConfig(): Observable<boolean> {
    return new Observable<boolean>(subscriber => {
      this.configScannerService.query({size: 1000}).subscribe(
        (res: HttpResponse<any>) => {
          this.configs = res.body;
          this.loadingScannersConfig = false;
          subscriber.next(true);
        },
        (res: HttpResponse<any>) => {
          this.loadingScannersConfig = false;
          subscriber.next(false);
        }
      );
    });
  }

  resolveRequestParams(): any {
    const req = {
      taskId: null,
      alterable: this.taskForm.get('alterable').value,
      comment: this.taskForm.get('comment').value,
      config: {
        id: this.taskForm.get('configId').value,
      },
      hostsOrdering: this.taskForm.get('hostsOrdering').value,
      name: this.taskForm.get('name').value,
      scanner: {
        id: this.taskForm.get('scannerId').value
      },
      schedule: null,
      schedulePeriods: 0,
      target: {
        id: this.taskForm.get('targetId').value,
      }
    };
    if (this.taskForm.get('scheduleId').value !== null && this.taskForm.get('scheduleId').value !== '') {
      req.schedule = {
        id: this.taskForm.get('scheduleId').value
      };
    }
    if (this.task) {
      req.taskId = this.task.id;
    }
    return req;
  }

  createTask() {
    this.creating = true;
    this.taskService.create(this.resolveRequestParams()).subscribe(target => {
      this.utmToastService.showSuccessBottom('Task created successfully');
      this.activeModal.close();
      this.taskCreated.emit(target.body.id);
    }, error1 => {
      this.creating = false;
      this.utmToastService.showError('Error creating Task',
        error1.error.statusText);
    });
  }

  editTask() {
    this.creating = true;
    this.taskService.update(this.resolveRequestParams()).subscribe(target => {
      this.utmToastService.showSuccessBottom('Task edited successfully');
      this.activeModal.close();
      this.taskCreated.emit(target.body.id);
    }, error1 => {
      this.creating = false;
      this.utmToastService.showError('Error editing Task',
        error1.error.statusText);
    });
  }

  newTarget() {
    const modal = this.modalService.open(TargetCreateComponent, {centered: true});
    modal.componentInstance.targetCreated.subscribe(targetId => {
      this.taskForm.get('targetId').setValue(targetId);
      this.getTargets();
    });
  }

  newSchedule() {
    const modal = this.modalService.open(ScheduleCreateComponent, {centered: true});
    modal.componentInstance.scheduleCreated.subscribe(schedule => {
      this.taskForm.get('scheduleId').setValue(schedule);
      this.getSchedules();
      this.runOnce = false;
      this.changeScheduleValidity();
    });
  }

  changePeriod($event) {
    this.runOnce = !this.runOnce;
    this.changeScheduleValidity();
  }

  changeScheduleValidity() {
    if (this.runOnce) {
      this.taskForm.get('scheduleId').setValue(null);
      this.taskForm.get('scheduleId').clearValidators();
      this.taskForm.get('scheduleId').updateValueAndValidity();
      this.taskForm.updateValueAndValidity();
      this.taskForm.get('schedulePeriods').setValue(this.runOnce);
    } else {
      this.taskForm.get('schedulePeriods').setValue(null);
      this.taskForm.get('scheduleId').setValidators(Validators.required);
    }
  }

  setSchedule($event) {
    this.runOnce = false;
    this.changeScheduleValidity();
  }

  getScannerId() {
    const index = this.scanners.findIndex(value => value.name === 'System Discovery');
    if (index !== -1) {
      return this.scanners[index].uuid;
    } else {
      return null;
    }
  }

  private setSimpleScanForm() {
    this.taskForm.get('name').setValue('Immediate scan of network');
    this.taskForm.get('comment').setValue('');
    this.taskForm.get('targetId').setValue(null);
    this.taskForm.get('scheduleId').setValue(null);
    this.taskForm.get('schedulePeriods').setValue(null);
    this.taskForm.get('inAssets').setValue(true);
    this.taskForm.get('applyOverrides').setValue(true);
    this.taskForm.get('minQod').setValue(70);
    this.taskForm.get('alterable').setValue(false);
    this.taskForm.get('autoDelete').setValue(false);
    this.taskForm.get('autoDeleteData').setValue(false);
    this.taskForm.get('scannerId').setValue(this.getScannerId());

  }

  private setTaskForm() {
    this.taskForm.get('id').setValue(this.task.id);
    this.taskForm.get('name').setValue(this.task.name);
    this.taskForm.get('comment').setValue(this.task.comment);
    this.taskForm.get('targetId').setValue(this.task.target.uuid);
    this.taskForm.get('scheduleId').setValue(this.task.schedule.uuid);
    this.taskForm.get('schedulePeriods').setValue(this.task.schedule.period);
    this.taskForm.get('inAssets').setValue(this.extractFromPreferences('in_assets'));
    this.taskForm.get('applyOverrides').setValue(this.extractFromPreferences('assets_apply_overrides'));
    this.taskForm.get('minQod').setValue(this.extractFromPreferences('assets_min_qod'));
    this.taskForm.get('alterable').setValue(Number(this.task.alterable));
    this.taskForm.get('autoDelete').setValue(this.extractFromPreferences('auto_delete'));
    this.taskForm.get('autoDeleteData').setValue(this.extractFromPreferences('auto_delete_data'));
    this.taskForm.get('scannerId').setValue(this.task.scanner.uuid);
    this.taskForm.get('scannerType').setValue(this.task);
    this.taskForm.get('configId').setValue(this.task.config.uuid);
    // this.taskForm.get('hostsOrdering').setValue(this.task);
    this.taskForm.get('sourceIface').setValue(this.extractFromPreferences('source_iface'));
    this.taskForm.get('maxChecks').setValue(this.extractFromPreferences('max_checks'));
    this.taskForm.get('maxHosts').setValue(this.extractFromPreferences('max_hosts'));

    // disable properties unchangable
    if (this.task.alterable === 1) {
      this.taskForm.get('targetId').disable();
      this.taskForm.get('scannerId').disable();
      this.taskForm.get('configId').disable();
      this.taskForm.get('scannerType').disable();
    }
    if (this.task.schedule.uuid !== '') {
      this.runOnce = false;
      this.changeScheduleValidity();
    }
  }


}
