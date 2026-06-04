import {Component, Input, OnInit} from '@angular/core';
import {ONLINE_DOCUMENTATION_BASE} from '../../../../constants/global.constant';

@Component({
  selector: 'app-utm-online-documentation',
  templateUrl: './utm-online-documentation.component.html',
  styleUrls: ['./utm-online-documentation.component.css']
})
export class UtmOnlineDocumentationComponent implements OnInit {
  @Input() path: string;
  @Input() text: string;
  onlineDoc = ONLINE_DOCUMENTATION_BASE;

  constructor() {
  }

  ngOnInit() {
    this.onlineDoc = this.path ? this.onlineDoc + this.path : this.onlineDoc;
    this.text = this.text ? this.text : 'Questions? Check our documentation';
  }

}
