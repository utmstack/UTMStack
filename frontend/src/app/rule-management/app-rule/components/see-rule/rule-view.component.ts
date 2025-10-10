
import { Component, Input } from '@angular/core';
import { NgbActiveModal } from '@ng-bootstrap/ng-bootstrap';
import { Rule } from '../../../models/rule.model';
import * as yaml from 'js-yaml';

@Component({
  selector: 'app-rule-view',
  templateUrl: './rule-view.component.html',
  styleUrls: ['./rule-view.component.scss'],
})
export class RuleViewComponent {
 @Input() rowDocument: Rule;

  copied = false;

  get yamlString(): string {
    try {
      return yaml.dump(this.rowDocument, { indent: 2 });
    } catch (e) {
      return 'Error parsing YAML';
    }
  }

  copyToClipboard() {
    window.navigator['clipboard'].writeText(this.yamlString).then(() => {
      this.copied = true;
      setTimeout(() => (this.copied = false), 1500);
    });
  }
}



