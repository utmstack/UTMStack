import {HttpResponse} from '@angular/common/http';
import {Component, EventEmitter, OnInit, Output} from '@angular/core';
import {DataNatureTypeEnum} from '../../../../../shared/enums/nature-data.enum';
import {IndexPatternService} from '../../../../../shared/services/elasticsearch/index-pattern.service';
import {UtmIndexPattern} from '../../../../../shared/types/index-pattern/utm-index-pattern';
import {TimeFilterType} from '../../../../../shared/types/time-filter.type';
import {CHART_TYPES} from '../../../../shared/const/chart-type.const';

@Component({
  selector: 'app-visualization-filter',
  templateUrl: './visualization-filter.component.html',
  styleUrls: ['./visualization-filter.component.scss']
})
export class VisualizationFilterComponent implements OnInit {
  @Output() visFilterChange = new EventEmitter<any>();
  eventNatureTypes = [DataNatureTypeEnum.ALERT, DataNatureTypeEnum.EVENT, DataNatureTypeEnum.VULNERABILITY];
  filter = {};
  charts = CHART_TYPES;
  patterns: UtmIndexPattern[] = [];
  loadingPatterns = false;

  constructor(private indexPatternService: IndexPatternService) {
  }

  ngOnInit() {
    this.getIndexPatterns();
  }

  searchByText() {
    this.visFilterChange.emit(this.filter);
  }

  searchByCreation($event: TimeFilterType) {
    this.filter['createdDate.greaterThanOrEqual'] = $event.timeFrom;
    this.filter['createdDate.lessThanOrEqual'] = $event.timeTo;
    this.visFilterChange.emit(this.filter);
  }

  searchByLastModification($event: TimeFilterType) {
    this.filter['modifiedDate.greaterThanOrEqual'] = $event.timeFrom;
    this.filter['modifiedDate.lessThanOrEqual'] = $event.timeTo;
    this.visFilterChange.emit(this.filter);
  }

  getIndexPatterns() {
    const req = {
      page: 0,
      size: 1000
    };
    this.loadingPatterns = true;
    this.indexPatternService.query(req).subscribe(
      (res: HttpResponse<any>) => this.onSuccess(res.body, res.headers),
      (res: HttpResponse<any>) => this.onError(res.body)
    );
  }

  private onSuccess(data, headers) {
    this.patterns = data;
    this.loadingPatterns = false;
  }

  private onError(error) {
    // this.alertService.error(error.error, error.message, null);
  }

  onSearchFor($event: string) {
    this.filter['name.contains'] = $event;
    this.visFilterChange.emit(this.filter);
  }
}
