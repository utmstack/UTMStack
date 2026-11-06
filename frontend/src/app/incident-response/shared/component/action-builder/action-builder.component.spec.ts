import {NO_ERRORS_SCHEMA} from '@angular/core';
import {CommonModule} from '@angular/common';
import {async, ComponentFixture, TestBed} from '@angular/core/testing';
import {FormBuilder, ReactiveFormsModule} from '@angular/forms';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {NgSelectModule} from '@ng-select/ng-select';
import {Observable} from 'rxjs';

import {UtmNetScanService} from '../../../../assets-discover/shared/services/utm-net-scan.service';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {InputClassResolve} from '../../../../shared/util/input-class-resolve';
import {WorkflowActionsService} from '../../services/workflow-actions.service';
import {ActionBuilderComponent} from './action-builder.component';

describe('ActionBuilderComponent', () => {
  let component: ActionBuilderComponent;
  let fixture: ComponentFixture<ActionBuilderComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      // ReactiveFormsModule: [formGroup]/formControlName. NgSelectModule: the
      // ng-select controls carry formControlName, so they need their NG_VALUE_ACCESSOR
      // (NO_ERRORS_SCHEMA alone leaves the FormControlName with no accessor -> NG01203).
      // CommonModule: async pipe on the [items] observables.
      imports: [ReactiveFormsModule, NgSelectModule, CommonModule],
      declarations: [ActionBuilderComponent],
      // sibling UI (app-utm-toggle, app-formcontrol-error, app-action-conditional,
      // app-utm-code-view) is out of scope here
      schemas: [NO_ERRORS_SCHEMA],
      providers: [
        InputClassResolve,
        WorkflowActionsService,
        // UtmNetScanService pulls HttpClient + RefreshDataService; UtmToastService
        // pulls Toastr/Translate/Router - none exercised by create.
        // NEVER-emitting (not of({body:[]})) so the pipe's tap/finalize don't flip
        // noPlatforms/loadingPlatforms synchronously during ngOnInit -> NG0100
        // ExpressionChangedAfterItHasBeenChecked on the first change-detection pass.
        {provide: UtmNetScanService, useValue: {getAssetsPlatforms: () => new Observable<void>(() => {}), query: () => new Observable<void>(() => {})}},
        {provide: UtmToastService, useValue: {showError: () => {}}},
        {provide: NgbModal, useValue: {open: () => ({componentInstance: {}, result: {then: () => {}}})}}
      ]
    })
      .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ActionBuilderComponent);
    component = fixture.componentInstance;
    // template binds [formGroup] + formControlName and ngOnInit subscribes to
    // control valueChanges - provide the group before first change detection
    const fb = new FormBuilder();
    component.group = fb.group({
      agentPlatform: [''],
      agentType: [false],
      excludedAgents: [[]],
      defaultAgent: [''],
      shell: ['powershell'],
      actions: [[]],
      command: ['']
    });
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
