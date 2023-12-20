import {async, ComponentFixture, TestBed} from '@angular/core/testing';

import {UtmOnlineDocumentationComponent} from './utm-online-documentation.component';

describe('UtmOnlineDocumentationComponent', () => {
  let component: UtmOnlineDocumentationComponent;
  let fixture: ComponentFixture<UtmOnlineDocumentationComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ UtmOnlineDocumentationComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(UtmOnlineDocumentationComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
