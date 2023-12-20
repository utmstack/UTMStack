import {Component, Input, OnInit} from '@angular/core';
import {AlertTags} from '../../../../../shared/types/alert/alert-tag.type';

@Component({
  selector: 'app-alert-tag-label',
  templateUrl: './alert-tag-label.component.html',
  styleUrls: ['./alert-tag-label.component.scss']
})
export class AlertTagLabelComponent implements OnInit {
  @Input() tags: AlertTags[];


  constructor() {
  }

  ngOnInit() {
  }


}
