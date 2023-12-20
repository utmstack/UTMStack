import {Component, Input, OnInit} from '@angular/core';

@Component({
  selector: 'app-alert-tags',
  templateUrl: './alert-tags-view.component.html',
  styleUrls: ['./alert-tags-view.component.scss']
})
export class AlertTagsViewComponent implements OnInit {

  @Input() tags: string[] = [];
  background: string;

  constructor() {
  }

  ngOnInit() {
    this.background = 'border-slate-800 text-slate-800';
    // @ts-ignore
    if (this.tags === 'None') {
      this.tags = [];
    }
  }

}
