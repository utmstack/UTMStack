import {Component, OnInit} from '@angular/core';
import {ElasticSearchIndexService} from '../../shared/services/elasticsearch/elasticsearch-index.service';
import {FieldDataService} from '../../shared/services/elasticsearch/field-data.service';
import {ElasticSearchFieldInfoType} from '../../shared/types/elasticsearch/elastic-search-field-info.type';

@Component({
  selector: 'app-rule-field-browser',
  templateUrl: './rule-field-browser.component.html',
  styleUrls: ['./rule-field-browser.component.css']
})
export class RuleFieldBrowserComponent implements OnInit {
  fields: ElasticSearchFieldInfoType[] = [];
  fieldsOriginal: ElasticSearchFieldInfoType[] = [];
  loadingFields = true;
  searching = false;
  pattern: string;
  pageStart = 0;
  pageEnd = 70;

  constructor(private elasticSearchIndexService: ElasticSearchIndexService,
              private fieldDataService: FieldDataService) {
  }

  ngOnInit() {
  }

  getFields(indexPattern: string) {
    this.fieldDataService.getFields(indexPattern).subscribe(field => {
      this.fields = field;
      this.fieldsOriginal = field;
      this.loadingFields = false;
    });
  }

  onSearch($event: string) {
    this.searching = true;
    this.doSearch($event).then(() => this.searching = false);
  }

  doSearch(search: string): Promise<boolean> {
    return new Promise<boolean>(resolve => {
      if (search !== '') {
        this.fields = this.fieldsOriginal.filter(value => {
          return value.name.toLowerCase().includes(search.toLowerCase()) && !value.name.includes('.keyword');
        });
      } else {
        this.fields = this.fieldsOriginal.filter(value => {
          return !value.name.includes('.keyword');
        });
      }
      resolve(true);
    });
  }

  onScroll() {
    this.pageEnd += 50;
  }
}
