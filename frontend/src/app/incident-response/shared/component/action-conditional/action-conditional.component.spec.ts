import {NO_ERRORS_SCHEMA} from '@angular/core';
import {async, ComponentFixture, TestBed} from '@angular/core/testing';
import {NgbDropdownModule} from '@ng-bootstrap/ng-bootstrap';

import {ActionConditionalComponent} from './action-conditional.component';

describe('ActionConditionalComponent', () => {
  let component: ActionConditionalComponent;
  let fixture: ComponentFixture<ActionConditionalComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [NgbDropdownModule], // #dropTracker="ngbDropdown" is @ViewChild({static:true})
      declarations: [ActionConditionalComponent],
      schemas: [NO_ERRORS_SCHEMA]
    })
      .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ActionConditionalComponent);
    component = fixture.componentInstance;
    // template calls getActionConditionalLabel/Icon() which read this.option.key
    component.option = {key: 0 as any, value: ';'};
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
