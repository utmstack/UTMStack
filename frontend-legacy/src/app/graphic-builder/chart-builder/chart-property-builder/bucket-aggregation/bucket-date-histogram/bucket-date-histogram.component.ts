import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';
import {BucketDateHistogramType} from '../../../../../shared/chart/types/metric/bucket-date-histogram.type';
import {BucketDateHistogramInterval, TIME_INTERVAL} from '../../../../../shared/constants/time-interval.const';
import {InputClassResolve} from '../../../../../shared/util/input-class-resolve';
import {getErrorCountForm} from '../../shared/functions/form-error-count';

@Component({
  selector: 'app-bucket-date-histogram',
  templateUrl: './bucket-date-histogram.component.html',
  styleUrls: ['./bucket-date-histogram.component.scss']
})
export class BucketDateHistogramComponent implements OnInit {
  intervals: BucketDateHistogramInterval[] = TIME_INTERVAL;
  showCustom = false;
  formDateHistogram: FormGroup;
  @Output() dateHistogramChange = new EventEmitter<BucketDateHistogramType>();
  @Output() errorsCount = new EventEmitter<number>();
  @Input() mode: string;
  @Input() dateHistogram: BucketDateHistogramType;

  constructor(private fb: FormBuilder, public inputClass: InputClassResolve) {
  }

  ngOnInit() {
    this.initFormDateHistogram();
    if (this.dateHistogram) {
      this.formDateHistogram.patchValue(this.dateHistogram);
    }
    this.formDateHistogram.valueChanges.subscribe((value => {
      this.dateHistogramChange.emit(value);
      this.errorsCount.emit(this.countErrors());
    }));
  }

  initFormDateHistogram() {
    this.formDateHistogram = this.fb.group(
      {
        interval: [[], Validators.required]
      });
  }

  addCustomValue($event: any) {
    if (['s', 'm', 'h', 'd', 'd', 'M'].includes($event.target.value)) {
      this.intervals.push($event.target.value);
    }
  }

  countErrors(): number {
    return getErrorCountForm(this.formDateHistogram);
  }
}
