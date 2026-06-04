import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';

@Component({
  selector: 'app-utm-mail-list',
  templateUrl: './utm-mail-list.component.html',
  styleUrls: ['./utm-mail-list.component.scss']
})
export class UtmMailListComponent implements OnInit {
  @Input() mails: string;
  @Output() emailChange = new EventEmitter<string>();
  mailList: string[] = [];
  emailValid: boolean;

  constructor() {
  }

  ngOnInit() {
    if (this.mails) {
      this.mailList = this.mails.split(',');
    }
  }

  validEmail(email: string) {
    // tslint:disable-next-line:max-line-length
    const regexp = new RegExp(/^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/);
    this.emailValid = regexp.test(email);
  }

  addEmail(email: string) {
    this.mailList.push(email);
    this.emailValid = undefined;
    this.emailChange.emit(this.mailList.toString());
  }


  deleteEmail(index) {
    this.mailList.splice(index, 1);
    this.emailChange.emit(this.mailList.toString());
  }

  mailHasAdded(email: string): boolean {
    return this.mailList.findIndex(control => control === email) !== -1;
  }

}
