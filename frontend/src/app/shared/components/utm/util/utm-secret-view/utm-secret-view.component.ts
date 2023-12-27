import {Component, Input, OnChanges, OnInit, SimpleChanges} from '@angular/core';
import {UtmToastService} from '../../../../alert/utm-toast.service';

@Component({
  selector: 'app-utm-secret-view',
  templateUrl: './utm-secret-view.component.html',
  styleUrls: ['./utm-secret-view.component.scss']
})
export class UtmSecretViewComponent implements OnInit, OnChanges {
  @Input() secret: string;
  @Input() allowCopy = true;
  private copy: any;
  str: string;
  view = false;

  constructor(private utmToastService: UtmToastService) {
  }

  ngOnInit() {
    this.str = this.secret;
    this.secret = this.maskSecrets(this.secret);
  }

  ngOnChanges(changes: SimpleChanges) {
    if (changes && changes.secret && changes.secret.firstChange === false && changes.secret.currentValue) {
      const currentValue = changes.secret.currentValue;
      this.str = currentValue;
      this.secret = this.maskSecrets(currentValue);
    }
  }

  copyCode() {
    const selBox = document.createElement('textarea');
    const copyText = this.removeSecretsTags(this.str.split('<br>').join(''));
    selBox.style.position = 'fixed';
    selBox.style.left = '0';
    selBox.style.top = '0';
    selBox.style.opacity = '0';
    selBox.value = copyText;
    document.body.appendChild(selBox);
    selBox.focus();
    selBox.select();
    document.execCommand('copy');

    document.body.removeChild(selBox);
    this.utmToastService.showInfo('Copied', 'Secret has been copy to clipboard');
  }


  maskSecrets(str: string) {
    return '*'.repeat(str.length);
  }

  removeSecretsTags(str) {
    const regex = /<\/?secret>/g;
    return str.replace(regex, '');
  }

  toggleViewSecret() {
    this.view = !this.view;
  }

  getSecret() {
    if (this.view) {
      return this.str;
    } else {
      return this.secret;
    }
  }
}
