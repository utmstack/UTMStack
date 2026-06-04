import {Component, Input, OnInit} from '@angular/core';

@Component({
  selector: 'app-alert-impact',
  templateUrl: './alert-impact.component.html',
  styleUrls: ['./alert-impact.component.scss']
})
export class AlertImpactComponent implements OnInit {
  @Input() impact: { [key: string]: number};
  background: string;
  impactFields: { id: string; value: number }[];

  constructor() {
  }

  ngOnInit() {
    this.impactFields = this.getFields(this.impact);
  }

  getFields(obj: any, prefix = ''): { id: string; value: number }[] {
    const fields: { id: string, value: number}[] = [];

    for (const key in obj) {
      if (obj.hasOwnProperty(key)) {
            fields.push({
              id: key,
              value: obj[key]
            });
      }
    }
    return fields;
  }
}
