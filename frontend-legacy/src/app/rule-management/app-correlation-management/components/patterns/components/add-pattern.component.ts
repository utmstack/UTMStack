import {Component, Input, OnInit} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../../shared/alert/utm-toast.service';
import {RegexPattern} from '../../../../models/rule.model';
import {PatternManagerService} from '../../../services/pattern-manager.service';

@Component({
  selector: 'app-add-pattern',
  templateUrl: './add-pattern.component.html',
  styleUrls: ['./add-pattern.component.scss']
})
export class AddPatternComponent implements OnInit {

  @Input() pattern: RegexPattern;
  patternForm: FormGroup;
  loading = false;
  mode: 'ADD' | 'EDIT';

  constructor(public activeModal: NgbActiveModal,
              private patternManagerService: PatternManagerService,
              private formBuilder: FormBuilder,
              private utmToastService: UtmToastService) {
  }

  ngOnInit() {
    this.mode = this.pattern ? 'EDIT' : 'ADD';
    this.initForm();
  }

  get patternId() {
    return this.patternForm.get('patternId');
  }

  get patternDefinition() {
    return this.patternForm.get('patternDefinition');
  }

  get patternDescription() {
    return this.patternForm.get('patternDescription');
  }


  initForm() {
    this.patternForm = this.formBuilder.group({
      patternId: [ this.pattern ? this.pattern.patternId : '', [Validators.required, Validators.maxLength(250)]],
      patternDefinition: [ this.pattern ? this.pattern.patternDefinition : '', [Validators.required, Validators.maxLength(250)]],
      patternDescription: [this.pattern ? this.pattern.patternDescription : '']
    });
  }

  onSubmit() {
    if (this.patternForm.valid) {
      this.loading = true;
      const formPattern: RegexPattern = this.patternForm.value;
      const patternToSave = this.mode === 'ADD' ? formPattern : {
        ...this.pattern,
        ...formPattern
      };
      this.patternManagerService
        .saveRegexPattern(this.mode, patternToSave)
           .subscribe({
             next: response => {
               this.patternForm.reset();
               this.loading = false;
               this.utmToastService.showSuccessBottom(this.mode === 'ADD'
                       ? 'Pattern regex saved successfully' : 'Pattern regex edited successfully');
               this.activeModal.close(true);
             },
             error: err => {
               this.loading = false;
               this.utmToastService.showError('Error', this.mode === 'ADD'
                   ? 'Error saving pattern' : 'Error editing pattern');
             }
           });
    }
  }

}
