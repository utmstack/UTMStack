import {Component, EventEmitter, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';
import {BucketTermsType} from '../../../../../shared/chart/types/metric/bucket-terms.type';
import {InputClassResolve} from '../../../../../shared/util/input-class-resolve';
import {getErrorCountForm} from '../../shared/functions/form-error-count';

@Component({
  selector: 'app-bucket-significant-terms',
  templateUrl: './bucket-significant-terms.component.html',
  styleUrls: ['./bucket-significant-terms.component.scss']
})
export class BucketSignificantTermsComponent implements OnInit {
  @Output() termSigBucketChange = new EventEmitter<BucketTermsType>();
  @Output() errorsCount = new EventEmitter<number>();
  formBucketSigTerms: FormGroup;

  constructor(private fb: FormBuilder,
              public inputClass: InputClassResolve) {
  }

  ngOnInit() {
    this.initFormBucketTerms();
    this.termSigBucketChange.emit(this.formBucketSigTerms.value);
    this.formBucketSigTerms.valueChanges.subscribe(() => {
      this.termSigBucketChange.emit(this.formBucketSigTerms.value);
      this.errorsCount.emit(this.countErrors());
    });
  }

  initFormBucketTerms() {
    this.formBucketSigTerms = this.fb.group({
      size: [10, Validators.required],
    });
  }

  countErrors(): number {
    return getErrorCountForm(this.formBucketSigTerms);
  }

}
