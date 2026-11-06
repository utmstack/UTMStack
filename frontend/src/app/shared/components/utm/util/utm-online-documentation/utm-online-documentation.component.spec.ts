import {async, ComponentFixture, TestBed} from '@angular/core/testing';

import {SafePipe} from '../../../../pipes/safe.pipe';
import {UtmOnlineDocumentationComponent} from './utm-online-documentation.component';

describe('UtmOnlineDocumentationComponent', () => {
  let component: UtmOnlineDocumentationComponent;
  let fixture: ComponentFixture<UtmOnlineDocumentationComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [UtmOnlineDocumentationComponent, SafePipe] // [href]="onlineDoc|safe:'url'"
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
