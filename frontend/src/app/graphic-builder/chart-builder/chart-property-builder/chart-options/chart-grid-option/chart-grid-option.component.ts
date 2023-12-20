import {Component, EventEmitter, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup} from '@angular/forms';
import {Grid} from '../../../../../shared/chart/types/charts/chart-properties/grid/grid';
import {GRID_POSITION_LEFT, GRID_POSITION_TOP} from '../../../../shared/const/chart/grid-properties.const';

@Component({
  selector: 'app-chart-grid-option',
  templateUrl: './chart-grid-option.component.html',
  styleUrls: ['./chart-grid-option.component.scss']
})
export class ChartGridOptionComponent implements OnInit {
  // @Input() gridOption: Grid;
  @Output() gridOptionChange = new EventEmitter<Grid>();
  formGridOption: FormGroup;
  gridPositionsTop = GRID_POSITION_TOP;
  gridPositionsLeft = GRID_POSITION_LEFT;
  viewGrid = false;
  useIcon = false;

  constructor(private fb: FormBuilder) {
  }


  ngOnInit() {
    this.initFormPie();
    // if (this.gridOption) {
    //   this.formGridOption.patchValue(this.gridOption);
    // }
    this.formGridOption.valueChanges.subscribe(value => {
      this.gridOptionChange.emit(this.formGridOption.value);
    });
    this.gridOptionChange.emit(this.formGridOption.value);
  }

  initFormPie() {
    this.formGridOption = this.fb.group({
      top: [],
      right: [],
      left: [],
      bottom: [],
    });
  }

  viewGridProperties() {
    this.viewGrid = this.viewGrid ? false : true;
  }

}
