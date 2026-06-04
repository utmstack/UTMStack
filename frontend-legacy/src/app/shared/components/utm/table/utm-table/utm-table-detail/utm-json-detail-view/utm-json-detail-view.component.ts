import {Component, Input, OnInit} from '@angular/core';

@Component({
  selector: 'app-utm-json-detail-view',
  templateUrl: './utm-json-detail-view.component.html',
  styleUrls: ['./utm-json-detail-view.component.scss']
})
export class UtmJsonDetailViewComponent implements OnInit {
  @Input() rowDocument: any;
  @Input() errors: Record<string,string[]> = {};
  detailWidth: number;
  keys:string[]
  separatedObject:any[]=[]

  constructor() {
    this.detailWidth = window.innerWidth - 330;
  }

  ngOnInit() {
    this.keys = Object.keys(this.rowDocument);
    this.separatedObject = this.keys.map(key=>({[key]:this.rowDocument[key]}))
  }
  hasError(key: string): boolean {
    return !!this.errors[key];
  }

  getErrors(key: string): string[] {
    return this.errors[key] || [];
  }

}
