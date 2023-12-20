import {async, ComponentFixture, TestBed} from '@angular/core/testing';

import {FileHostOsComponent} from './file-host-os.component';

describe('FileHostOsComponent', () => {
  let component: FileHostOsComponent;
  let fixture: ComponentFixture<FileHostOsComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [FileHostOsComponent]
    })
      .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(FileHostOsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
