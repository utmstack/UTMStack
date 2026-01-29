import {HttpResponse} from '@angular/common/http';
import {Component, OnDestroy, OnInit} from '@angular/core';
import {FormBuilder, FormGroup} from '@angular/forms';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {forkJoin, from, of} from 'rxjs';
import {catchError, concatMap, finalize, map, tap, toArray} from 'rxjs/operators';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {AddRuleStepEnum, Mode, Rule, Status} from '../../../models/rule.model';
import {DataTypeService} from '../../../services/data-type.service';
import {RuleService} from '../../../services/rule.service';
import {ImportRuleService} from './import-rule.service';


interface RuleList {
    rule: Rule;
    valid: boolean;
    status: Status;
    errors: Record<string, string[]>;
    isLoading: boolean;
}

@Component({
  selector: 'app-import-rule',
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
  rules: RuleList[] = [];

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
      concatMap(ruleData => {
        const rule = ruleData.rule as Rule;
        return this.ruleService.saveRule('ADD', rule).pipe(
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
        );
      }),
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
      case 2: this.currentStep = AddRuleStepEnum.STEP2;
              break;
    }
  }

  validRules() {
    if (this.files.length > 0) {
      this.loading = true;
      const filesWithDataTypes = this.files.map(file => {
          return {
            ...file,
            dataTypes: file.dataTypes && file.dataTypes.length > 0 ? file.dataTypes : []
          };
      });

      // Fetch and filter data types for each file
      forkJoin(
          filesWithDataTypes.map(file =>
            forkJoin(
              file.dataTypes.map((dt: string) =>
                this.dataTypeService.getAll({ search: dt })
                  .pipe(
                      map(res => {
                        const dataTypes =  res.body;

                        return dataTypes.find(d => d.dataType === dt);
                      })
                  )
              )
            ).pipe(
              map(filteredDataTypes => ({
                ...file,
                confidentiality: file.impact.confidentiality || 0,
                integrity: file.impact.integrity || 0,
                availability: file.impact.availability || 0,
                definition: file.where || '',
                afterEvents: file.afterEvents || file.correlation || [],
                dataTypes: filteredDataTypes.filter(dt => !!dt)
              }))
            ),
          ),
        ).pipe(finalize(() => (this.loading = false))).subscribe(updatedFiles => {
          this.rules = updatedFiles.map(file => {
            let rule: Rule = {
                ...file,
                dataTypes: file.dataTypes.length > 0 ? file.dataTypes : [],
            };
            const {isValid, errors} = this.importRuleService.isValidRule(rule);

            Object.keys(rule).forEach(key => {
              if (rule[key] === null) {
                rule = {[key]: null, ...rule};
              }
            });

            return {
              rule,
              valid: isValid,
              status: isValid ? ('valid' as Status) : ('error' as Status),
              isLoading: false,
              errors
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


