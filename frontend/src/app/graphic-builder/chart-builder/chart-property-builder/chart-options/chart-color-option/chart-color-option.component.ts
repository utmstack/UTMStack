import {Component, EventEmitter, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup} from '@angular/forms';
import {Color} from '../../../../../shared/chart/types/charts/chart-properties/color/color';
import {UTM_COLOR_THEME} from '../../../../../shared/constants/utm-color.const';

@Component({
  selector: 'app-chart-color-option',
  templateUrl: './chart-color-option.component.html',
  styleUrls: ['./chart-color-option.component.scss']
})
export class ChartColorOptionComponent implements OnInit {
  @Output() colorOptionChange = new EventEmitter<Color | string[]>();
  formColorOption: FormGroup;
  viewColors = false;

  constructor(private fb: FormBuilder) {
  }

  ngOnInit() {
    this.initFormPie();
    this.formColorOption.valueChanges.subscribe(value => {
      this.colorOptionChange.emit(this.formColorOption.get('color').value);
    });
    this.colorOptionChange.emit(this.formColorOption.get('color').value);
  }

  initFormPie() {
    this.formColorOption = this.fb.group({
      color: [UTM_COLOR_THEME]
    });
  }


  viewColorsProperties() {
    this.viewColors = this.viewColors ? false : true;
  }

}
