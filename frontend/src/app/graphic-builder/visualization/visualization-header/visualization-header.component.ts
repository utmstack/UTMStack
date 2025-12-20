import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UTM_CHART_ICONS} from '../../../shared/constants/icons-chart.const';
import {UtmIndexPattern} from '../../../shared/types/index-pattern/utm-index-pattern';

@Component({
  selector: 'app-visualization-header',
  templateUrl: './visualization-header.component.html',
  styleUrls: ['./visualization-header.component.scss']
})
export class VisualizationHeaderComponent implements OnInit {
  @Input() chartType: string;
  @Output() indexPatternInitialized = new EventEmitter<string[]>();
  @Output() indexPatternSelected = new EventEmitter<UtmIndexPattern>();
  @Output() cancelled = new EventEmitter<void>();
  @Output() saved = new EventEmitter<void>();
  @Input() sqlMode = false;
  @Output() sqlModeToggled = new EventEmitter<boolean>();
  @Input() pattern: UtmIndexPattern;

  constructor(public activeModal: NgbActiveModal) {
  }

  ngOnInit() {

  }

  indexPatternChange(pattern: UtmIndexPattern) {
    this.indexPatternSelected.emit(pattern);
  }

  indexPatternLoaded(indexPatternNames: string[]) {
    this.indexPatternInitialized.emit(indexPatternNames);
  }

  cancel() {
    this.cancelled.emit();
  }

  save() {
    this.saved.emit();
  }

  toggleSqlMode() {
    this.sqlMode = !this.sqlMode;
    this.sqlModeToggled.emit(this.sqlMode);
  }

  chartIconResolver(): string {
    return this.chartType ? UTM_CHART_ICONS[this.chartType] : '';
  }
}
