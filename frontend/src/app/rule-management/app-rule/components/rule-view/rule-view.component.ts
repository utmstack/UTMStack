import {Component, Input} from '@angular/core';
import * as yaml from 'js-yaml';
import {mapRuleToYaml} from '../../../../shared/components/utm/util/utm-file-upload/shared/rule-yaml.mapper';
import {Rule} from '../../../models/rule.model';

@Component({
  selector: 'app-rule-view',
  templateUrl: './rule-view.component.html',
  styleUrls: ['./rule-view.component.scss'],
})
export class RuleViewComponent {
  @Input() rowDocument: Rule;

  copied = false;
  editorOptions: any = {
    theme: 'vs-light',
    language: 'yaml',
    automaticLayout: true,
    fontSize: 13,
    lineHeight: 24,
    fontFamily: 'Courier New, Monaco, Menlo, monospace',
    lineNumbers: 'on',
    wordWrap: 'on',
    minimap: { enabled: false },
    scrollBeyondLastLine: false,
    formatOnPaste: true,
    formatOnType: true,
    tabSize: 2,
    insertSpaces: true,
    readOnly: true,
    scrollbar: {
      vertical: 'auto',
      horizontal: 'hidden',
      useShadows: true,
      verticalSliderSize: 12,
      horizontalSliderSize: 0
    }
  };

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
