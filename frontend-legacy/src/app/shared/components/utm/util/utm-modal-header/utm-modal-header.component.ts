import {AfterViewChecked, AfterViewInit, ChangeDetectorRef, Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';

@Component({
  selector: 'app-utm-modal-header',
  templateUrl: './utm-modal-header.component.html',
  styleUrls: ['./utm-modal-header.component.scss']
})
export class UtmModalHeaderComponent implements OnInit, AfterViewChecked, AfterViewInit {
  @Input() name: string;
  @Input() type: 'default' | 'delete' | 'warning' | 'success' = 'default';
  @Input() icon?: string;
  @Input() showCloseButton = true;
  @Output() closeModal = new EventEmitter<boolean>();


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
