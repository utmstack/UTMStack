import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormArray, FormBuilder, FormGroup, Validators} from '@angular/forms';
import {UtmToastService} from '../../../shared/alert/utm-toast.service';
import {AdNotificationConfService} from '../shared/services/ad-notification-conf.service';
import {AdNotificationConfType} from '../shared/type/ad-notification-conf.type';

@Component({
  selector: 'app-ad-notifications-config-create',
  templateUrl: './ad-notifications-config-create.component.html',
  styleUrls: ['./ad-notifications-config-create.component.scss']
})
export class AdNotificationsConfigCreateComponent implements OnInit {
  @Input() notifyConfig: any;
  @Output() notificationCreated = new EventEmitter<string>();
  creating = false;
  adNotification: AdNotificationConfType = {
    mailList: '',
    mailStatus: true,
    smsList: '',
    smsStatus: true
  };
  mailList: string[] = [];
  phoneList: string[] = [];
  emailValid: boolean;
  invalidPhone: boolean;
  phoneNumber = '';
  dialCode = '+1';
  formConfig: FormGroup;

  constructor(private adNotificationConfService: AdNotificationConfService,
              private fb: FormBuilder,
              private utmToastService: UtmToastService) {
  }

  get mailListArray() {
    return this.formConfig.get('mailList') as FormArray;
  }

  get phoneListArray() {
    return this.formConfig.get('mailList') as FormArray;
  }

  ngOnInit() {
    this.loadConfig();
    this.initForm();
  }

  initForm() {
    this.formConfig = this.fb.group({
      mailList: this.fb.array([]),
      phoneList: this.fb.array([]),
    });
  }

  addMailToList() {
    this.mailListArray.push(this.fb.group({
      mail: ['', [Validators.required, Validators.email]]
    }));
  }

  deleteFromMailList(index) {
    this.mailListArray.removeAt(index);
  }

  addPhoneToList() {
    this.phoneListArray.push(this.fb.group({
      phone: ['', [Validators.required]]
    }));
  }


  loadConfig() {
    this.adNotificationConfService.query().subscribe(notify => {
      if (notify.body.length > 0) {
        this.adNotification = notify.body[0];
        this.mailList = this.adNotification.mailList.split(',');
        this.phoneList = this.adNotification.smsList.split(',');
      }
    });
  }

  validEmail(email: string) {
    // tslint:disable-next-line:max-line-length
    const regexp = new RegExp(/^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/);
    this.emailValid = regexp.test(email);
  }

  addEmail(email: string) {
    this.mailList.push(email);
    this.emailValid = undefined;
  }

  deleteEmail(index) {
    this.mailList.splice(index, 1);
  }

  addPhone() {
    this.phoneList.push(this.phoneNumber);
    this.phoneNumber = '';
  }

  deletePhone(index) {
    this.phoneList.splice(index, 1);
  }


  createNotifyConfig() {
    this.creating = true;
    this.adNotification.mailList = this.mailList.length === 0 ? null : this.mailList.toString();
    this.adNotification.smsList = this.phoneList.length === 0 ? null : this.phoneList.toString();

    this.adNotificationConfService.create(this.adNotification).subscribe(() => {
      this.notificationCreated.emit('notification created');
      this.utmToastService.showSuccessBottom('Tracker notification created successfully');
    }, error1 => {
      this.creating = false;
      this.utmToastService.showError('Error creating tracker notification ',
        'Error creating configuration, check your network');
    });
  }

  mailHasAdded(email: string): boolean {
    return this.mailList.findIndex(control => control === email) !== -1;
  }

  saveNotifyConfig() {
    this.creating = true;

    this.adNotification.mailList = this.mailList.length === 0 ? null : this.mailList.toString();
    this.adNotification.smsList = this.phoneList.length === 0 ? null : this.phoneList.toString();

    this.adNotificationConfService.update(this.adNotification).subscribe(() => {
      this.notificationCreated.emit('notification created');
      this.utmToastService.showSuccessBottom('Tracker notification updated successfully');
      this.creating = true;
    }, error1 => {
      this.creating = false;
      this.utmToastService.showError('Error updating tracker notification ',
        'Error updating configuration, check your network');
    });
  }

  hasErrorPhone($event: boolean) {
    this.invalidPhone = $event;
  }

  telInputObject($event: any) {
  }

  onCountryChange($event: any) {
    this.dialCode = '+' + $event.dialCode;
  }

  phoneHasAdded(): boolean {
    return this.phoneList.findIndex(value => value === this.phoneNumber) !== -1;
  }

  getNumber($event: any) {
    this.phoneNumber = $event;
  }

  deleteConfig() {
    this.adNotificationConfService.delete(this.adNotification.id).subscribe(() => {
      this.notificationCreated.emit('notification created');
      this.utmToastService.showSuccessBottom('Tracker notification deleted successfully');
      this.creating = true;
    }, error1 => {
      this.creating = false;
      this.utmToastService.showError('Error deleting tracker notification ',
        'Error updating configuration, check your network');
    });
  }
}
