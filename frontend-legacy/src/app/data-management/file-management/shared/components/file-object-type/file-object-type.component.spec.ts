import {async, ComponentFixture, TestBed} from '@angular/core/testing';

import {FileObjectTypeComponent} from './file-object-type.component';

describe('FileObjectTypeComponent', () => {
  let component: FileObjectTypeComponent;
  let fixture: ComponentFixture<FileObjectTypeComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [FileObjectTypeComponent]
    })
      .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(FileObjectTypeComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
