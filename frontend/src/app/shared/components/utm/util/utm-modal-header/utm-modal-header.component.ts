import {AfterViewChecked, AfterViewInit, ChangeDetectorRef, Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';

@Component({
  selector: 'app-utm-modal-header',
  templateUrl: './utm-modal-header.component.html',
  styleUrls: ['./utm-modal-header.component.scss']
})
export class UtmModalHeaderComponent implements OnInit, AfterViewChecked, AfterViewInit {
  @Input() name: string;
  @Input() type: 'delete' | 'normal';
  @Output() closeModal = new EventEmitter<boolean>();
  @Input() showCloseButton = true;

  constructor(public activeModal: NgbActiveModal,
              private cdr: ChangeDetectorRef) {
  }

  ngOnInit() {
  }

  ngAfterViewChecked(): void {
    this.cdr.detectChanges();
  }

  ngAfterViewInit() {
    this.cdr.detectChanges();
  }

}
