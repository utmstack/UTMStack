import {HttpResponse} from '@angular/common/http';
import {Component, OnDestroy, OnInit} from '@angular/core';
import {FormBuilder, FormGroup} from '@angular/forms';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {forkJoin, from, of} from 'rxjs';
import {catchError, concatMap, map, tap, toArray} from 'rxjs/operators';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {AddRuleStepEnum, Mode, Rule, Status} from '../../../models/rule.model';
import {DataTypeService} from '../../../services/data-type.service';
import {RuleService} from '../../../services/rule.service';
import {ImportRuleService} from './import-rule.service';

@Component({
  selector: 'app-add-rule',
  templateUrl: './import-rule.component.html',
  styleUrls: ['./import-rule.component.scss'],
})
export class ImportRuleComponent implements OnInit, OnDestroy {
  RULE_FORM = AddRuleStepEnum;
  ruleForm: FormGroup;
  mode: Mode = 'ADD';
  isSubmitting = false;
  savedVariables = [];
  rule: Rule;
  loading = false;
  currentStep: AddRuleStepEnum;
  stepCompleted: number[] = [];
  files = [];
  rules: Rule[] = [];

  constructor(private importRuleService: ImportRuleService,
              private dataTypeService: DataTypeService,
              private ruleService: RuleService,
              private utmToastService: UtmToastService,
              public activeModal: NgbActiveModal) {
  }

  ngOnInit() {
    this.currentStep = this.mode !== 'IMPORT' ? AddRuleStepEnum.STEP1 : AddRuleStepEnum.STEP0;
  }

  saveRule() {
    this.loading = true;
    from(this.rules).pipe(
      concatMap(rule =>
        this.ruleService.saveRule('ADD', rule).pipe(
          tap(() => rule.isLoading = !rule.isLoading),
          map((response: HttpResponse<any>) => {
            if (response.status === 204) {
              rule.status = 'saved';
              rule.isLoading = false;
              return rule;
            } else {
              throw new Error('Unexpected response status');
            }
          }),
          catchError(error => {
            rule.isLoading = false;
            rule.status = 'error';
            return of(rule);
          })
        )
      ),
      toArray()
    ).subscribe({
      next: response => {
        const hasError = response.some(r => r.status === 'error');
        const successResponse = response.every(r => r.status === 'saved');

        this.loading = false;
        this.isSubmitting = false;
        // this.currentStep = this.RULE_FORM.STEP3;
        this.next();

        if (response.length === 1) {
          if (successResponse) {
            this.utmToastService.showSuccessBottom('Rule imported successfully');
          } else if (hasError) {
            this.utmToastService.showError('Import failed', 'The rule could not be imported.');
          }
        } else {
          if (successResponse) {
            this.utmToastService.showSuccessBottom('Rules imported successfully');
          } else if (hasError) {
            this.utmToastService.showError('Import completed with errors', 'Some rules failed to import.');
          } else {
            this.utmToastService.showWarning('Import partially successful', 'Some rules were not saved.');
          }
        }
      }
    });
  }


  next() {
    this.stepCompleted.push(this.currentStep);
    switch (this.currentStep) {
      case 0: this.currentStep = AddRuleStepEnum.STEP1;
              this.validRules();
              break;
      case 1: this.currentStep = AddRuleStepEnum.STEP2;
              break;
      case 2: this.currentStep = AddRuleStepEnum.STEP3;
              break;
    }
  }

  validRules() {
    if (this.files.length > 0) {
      const filesWithDataTypes = this.files.map(file => {
          return {
            ...file,
            dataTypes: file.dataTypes && file.dataTypes.length > 0 ? file.dataTypes : []
          };
      });
      forkJoin(
          filesWithDataTypes.map(file =>
            forkJoin(
              file.dataTypes.map((dt: string) =>
                this.dataTypeService.getAll({ search: dt }).pipe(
                  map(res => res.body.length > 0 ? res.body[0] : null)
                )
              )
            ).pipe(
              map(filteredDataTypes => ({
                ...file,
                dataTypes: filteredDataTypes.filter(dt => !!dt)
              }))
            )
          )
        ).subscribe(updatedFiles => {
          this.rules = updatedFiles.map(file => ({
            ...file,
            dataTypes: file.dataTypes.length > 0 ? file.dataTypes : []
          }));

          this.rules = this.rules.map(rule => {
          const isValid = this.importRuleService.isValidRule(rule);

          return {
            ...rule,
            valid: isValid,
            status: isValid ? ('valid' as Status) : ('error' as Status)
          };

        });

        });
    } else {
      this.mode = 'ERROR';
    }
  }



  back() {
    this.stepCompleted.pop();
    switch (this.currentStep) {
      case 2: this.currentStep = AddRuleStepEnum.STEP1;
              break;
      case 1: this.currentStep = AddRuleStepEnum.STEP0;
              break;
    }
  }

  isCompleted(step: number) {
    return this.stepCompleted.findIndex(value => value === step) !== -1;
  }

  onFileChange($event: any): void {
    this.files = $event;
    this.files = this.files.filter(file => !file.error);
  }

  deleteRule(i: number) {
    this.rules.splice(i, 1);
  }

  showRule(rule: Rule) {
    rule.showDetail = !rule.showDetail;
  }

  close() {
    this.activeModal.close(true);
  }

  ngOnDestroy() {
    this.dataTypeService.resetTypes();
  }
}
