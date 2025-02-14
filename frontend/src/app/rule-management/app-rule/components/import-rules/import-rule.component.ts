import {Component, OnDestroy, OnInit} from '@angular/core';
import {FormBuilder, FormGroup} from '@angular/forms';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {forkJoin, of} from 'rxjs';
import {catchError, map} from 'rxjs/operators';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {AddRuleStepEnum, Mode, Rule} from '../../../models/rule.model';
import {DataTypeService} from '../../../services/data-type.service';
import {RuleService} from '../../../services/rule.service';
import {ImportRuleService} from "./import-rule.service";
import {HttpResponse} from "@angular/common/http";

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
    forkJoin(
      this.rules.map(rule =>
        this.ruleService.saveRule('ADD', rule).pipe(
          map((response: HttpResponse<any>) => {
            if (response.status === 204) {
              rule.status = true;
              return rule;
            } else {
              throw new Error('Unexpected response status');
            }
          }),
          catchError(error => {
            rule.status = false;
            return of(null);
          })
        )
      )
    ).subscribe({
      next: response => {
        this.isSubmitting = false;
        this.utmToastService.showSuccessBottom( 'Rule(s) imported successfully');
        this.activeModal.close(true);
      },
      error: err => {
        this.isSubmitting = false;
        this.utmToastService.showError('Error', 'Error importing rules');
        console.error('Error saving rule:', err.message);
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

          this.rules = this.rules.map(rule => ({
            ...rule,
            valid: this.importRuleService.isValidRule(rule)
        }));
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

  isValidStep(step: number) {
    switch (step) {
      case AddRuleStepEnum.STEP1:


      case AddRuleStepEnum.STEP2:
        return this.ruleForm.get('definition').valid;
    }
  }

  onFileChange($event: any): void {
    this.files = $event;
    this.files = this.files.filter(file => !file.error);
  }

  ngOnDestroy() {
    this.dataTypeService.resetTypes();
  }

  deleteRule(i: number) {
    this.rules.splice(i, 1);
  }

  showRule(rule: Rule) {
    rule.showDetail = !rule.showDetail;
  }
}
