import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ActionBuilderComponent } from './action-builder.component';

describe('ActionBuilderComponent', () => {
  let component: ActionBuilderComponent;
  let fixture: ComponentFixture<ActionBuilderComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ ActionBuilderComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ActionBuilderComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
