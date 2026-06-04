import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ActionConditionalComponent } from './action-conditional.component';

describe('ActionConditionalComponent', () => {
  let component: ActionConditionalComponent;
  let fixture: ComponentFixture<ActionConditionalComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ ActionConditionalComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ActionConditionalComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
