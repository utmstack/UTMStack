import {Component, Input, OnInit} from '@angular/core';
import {AbstractControl} from '@angular/forms';

@Component({
  selector: 'app-formcontrol-error',
  templateUrl: './formcontrol-error.component.html',
  styleUrls: ['./formcontrol-error.component.scss']
})
export class FormcontrolErrorComponent implements OnInit {
  @Input() formcontrol: AbstractControl;
  @Input() label: string;

  constructor() {
  }

  ngOnInit() {
  }

  getErrors(): string {
    let html = '';
    if (!this.formcontrol.valid) {
      Object.keys(this.formcontrol.errors).forEach((key) => {
        if (key === 'required') {
          if (this.label) {
            html += '<span>' + this.label + ' is required</span></br>';
          } else {
            html += '<span>This field is required</span></br>';
          }
        }
        if (key === 'pattern' || key === 'min' || key === 'max') {
          if (this.label) {
            html += '<span>' + this.label + ' is invalid</span></br>';
          } else {
            html += '<span>This field is invalid</span></br>';
          }
        }
      });
    }
    return html;
  }
}
