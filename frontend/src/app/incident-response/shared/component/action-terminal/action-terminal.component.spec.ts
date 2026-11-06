import {NO_ERRORS_SCHEMA} from '@angular/core';
import {async, ComponentFixture, TestBed} from '@angular/core/testing';
import {ReactiveFormsModule} from '@angular/forms';
import {NgbActiveModal, NgbPopoverModule} from '@ng-bootstrap/ng-bootstrap';

import {ActionTerminalComponent} from './action-terminal.component';

describe('ActionTerminalComponent', () => {
  let component: ActionTerminalComponent;
  let fixture: ComponentFixture<ActionTerminalComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [ReactiveFormsModule, NgbPopoverModule], // [formGroup] + [ngbPopover]
      declarations: [ActionTerminalComponent],
      // app-utm-modal-header / app-incident-variable-select are out of scope here
      schemas: [NO_ERRORS_SCHEMA],
      providers: [
        // NgbActiveModal is provided by NgbModal in the app; in the unit test the
        // component only reads it on close()/create(), so a stub suffices.
        {provide: NgbActiveModal, useValue: {close: () => {}, dismiss: () => {}}}
      ]
    })
      .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ActionTerminalComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
