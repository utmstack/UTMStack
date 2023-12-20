import {Component, EventEmitter, OnInit, Output} from '@angular/core';
import {ElasticFilterType} from '../../../../../shared/types/filter/elastic-filter.type';
import {ASSETS_FIELDS_FILTERS} from '../../../const/asset-field.const';

@Component({
  selector: 'app-assets-filter',
  templateUrl: './assets-filter.component.html',
  styleUrls: ['./assets-filter.component.css']
})
export class AssetsFilterComponent implements OnInit {
  @Output() filterChange = new EventEmitter<ElasticFilterType>();
  @Output() filterReset = new EventEmitter<boolean>();
  fieldFilters = ASSETS_FIELDS_FILTERS;

  constructor() {
  }

  ngOnInit() {
  }

  resetAllFilters() {
  }

  onFilterGenericChange($event: any) {
  }
}
