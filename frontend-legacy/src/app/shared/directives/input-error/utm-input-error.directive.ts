import {Directive, ElementRef, HostListener, Input, OnDestroy, OnInit} from '@angular/core';
import {NgControl} from '@angular/forms';
import {Subscription} from 'rxjs';

@Directive({
  selector: '[appUtmInputError]'
})
export class UtmInputErrorDirective implements OnInit, OnDestroy {
  @Input() label: string;
  statusChangeSubscription: Subscription;

  constructor(private elRef: ElementRef,
              private control: NgControl) {
  }

  ngOnInit(): void {
    this.statusChangeSubscription = this.control.statusChanges.subscribe(
      (status) => {
        if (status === 'INVALID') {
          this.showError();
        } else {
          this.removeError();
        }
      }
    );
  }

  ngOnDestroy(): void {
    this.statusChangeSubscription.unsubscribe();
  }

  @HostListener('blur', ['$event'])
  handleBlurEvent(event) {
    // This is needed to handle the case of clicking a required field and moving out.
    // Rest all are handled by status change subscription
    if (this.control.value == null || this.control.value === '') {
      if (this.control.errors) {
        this.showError();
      } else {
        this.removeError();
      }
    }
  }

  getErrors(): string {
    let html = '';
    if (!this.control.valid) {
      Object.keys(this.control.errors).forEach((key) => {
        if (key === 'required') {
          if (this.label) {
            html += '<span class="text-danger-300" id="errorIputSpan">' + this.label + ' is required</span>';
          } else {
            html += '<span class="text-danger-300" id="">This field is required</span>';
          }
        }
        if (key === 'pattern' || key === 'min') {
          if (this.label) {
            html += '<span class="text-danger-300" id="">' + this.label + ' is invalid</span>';
          } else {
            html += '<span class="text-danger-300" id="">This field is invalid</span>';
          }
        }
      });
    }
    return html;
  }

  private showError() {
    this.removeError();
    this.elRef.nativeElement.parentElement.insertAdjacentHTML('beforeend', this.getErrors());
  }

  private removeError(): void {
    const errorElement = document.getElementById('errorInputSpan');
    if (errorElement) {
      errorElement.remove();
    }
  }
}
