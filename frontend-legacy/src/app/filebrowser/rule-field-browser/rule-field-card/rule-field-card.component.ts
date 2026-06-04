import {Component, Input, OnInit} from '@angular/core';
import {UtmToastService} from '../../../shared/alert/utm-toast.service';
import {ElasticDataTypesEnum} from '../../../shared/enums/elastic-data-types.enum';
import {ElasticSearchFieldInfoType} from '../../../shared/types/elasticsearch/elastic-search-field-info.type';

@Component({
  selector: 'app-rule-field-card',
  templateUrl: './rule-field-card.component.html',
  styleUrls: ['./rule-field-card.component.css']
})
export class RuleFieldCardComponent implements OnInit {
  @Input() field: ElasticSearchFieldInfoType;
  viewCopyButton = false;

  constructor(private utmToastService: UtmToastService) {
  }

  ngOnInit() {
  }

  resolveIcon(): string {
    switch (this.field.type) {
      case ElasticDataTypesEnum.TEXT:
        return 'icon-text-color';
      case ElasticDataTypesEnum.LONG:
        return 'icon-hash';
      case ElasticDataTypesEnum.FLOAT:
        return 'icon-hash';
      case ElasticDataTypesEnum.DATE:
        return 'icon-calendar52';
      case ElasticDataTypesEnum.OBJECT:
        return 'icon-circle-css';
      case ElasticDataTypesEnum.BOOLEAN:
        return 'icon-split';
      default:
        return 'icon-question3';
    }
  }

  copyName(field: ElasticSearchFieldInfoType) {
    const selBox = document.createElement('textarea');
    selBox.style.position = 'fixed';
    selBox.style.left = '0';
    selBox.style.top = '0';
    selBox.style.opacity = '0';
    selBox.value = this.field.name;
    document.body.appendChild(selBox);
    selBox.focus();
    selBox.select();
    document.execCommand('copy');
    document.body.removeChild(selBox);
    this.utmToastService.showInfo('Field is in clipboard', this.field.name);
  }

  processField(name: string) {
    return name.replace('.keyword', '');
  }

}
