import {Component, Input, OnInit} from '@angular/core';
import * as yaml from 'js-yaml';
import {mapRuleToYaml} from '../../../../shared/components/utm/util/utm-file-upload/shared/rule-yaml.mapper';
import {Rule} from '../../../models/rule.model';

@Component({
  selector: 'app-rule-view',
  templateUrl: './rule-view.component.html',
  styleUrls: ['./rule-view.component.scss'],
})
export class RuleViewComponent implements OnInit {
  @Input() rowDocument: Rule;

  copied = false;

  ngOnInit() {
    console.log('Rule received:', this.rowDocument);
  }

  get yamlString(): string {
    try {
      const yamlModel = mapRuleToYaml(this.rowDocument);

      return yaml.dump(yamlModel, {
        indent: 2,
        lineWidth: -1,
        styles: {
          '!!null': 'empty'
        }
      });
    } catch (e) {
      return 'Error generating YAML';
    }
  }

  get yamlHighlighted(): string {
    return this.yamlString
      .replace(/^(\s*)([a-zA-Z0-9_]+):/gm, '$1<span class="yaml-key">$2</span>:')
      .replace(/: (.*)/g, ': <span class="yaml-value">$1</span>')
      .replace(/-\s+(.*)/g, '- <span class="yaml-value">$1</span>')
      .replace(/^\s{2,}(.+)/gm, '  <span class="yaml-value">$1</span>');
  }

  exportYaml() {
    const yamlContent = this.yamlString;

    const blob = new Blob([yamlContent], {
      type: 'text/yaml;charset=utf-8'
    });

    const url = window.URL.createObjectURL(blob);

    const a = document.createElement('a');
    a.href = url;

    a.download = `${this.rowDocument.name || 'rule'}.yml`
      .replace(/\s+/g, '-')
      .toLowerCase();
    a.click();

    window.URL.revokeObjectURL(url);
  }


  copyToClipboard() {
    (navigator as any).clipboard.writeText(this.yamlString).then(() => {
      this.copied = true;
      setTimeout(() => (this.copied = false), 1500);
    });
  }
}
