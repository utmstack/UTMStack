import {
  Component,
  ElementRef,
  EventEmitter,
  HostListener,
  Input,
  OnDestroy,
  OnInit,
  Output,
  ViewChild
} from '@angular/core';
import {FormArray, FormBuilder, FormGroup, Validators} from '@angular/forms';
import {ActivatedRoute, Router} from '@angular/router';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {debounceTime, filter, map, switchMap} from 'rxjs/operators';
import {NetScanType} from '../../assets-discover/shared/types/net-scan.type';
import {UtmToastService} from '../../shared/alert/utm-toast.service';
import {ElasticOperatorsEnum} from '../../shared/enums/elastic-operators.enum';
import {PrefixElementEnum} from '../../shared/enums/prefix-element.enum';
import {getValueFromPropertyPath} from '../../shared/util/get-value-object-from-property-path.util';
import {InputClassResolve} from '../../shared/util/input-class-resolve';
import {createElementPrefix, getElementPrefix} from '../../shared/util/string-util';
import {IncidentResponseRuleService} from '../shared/services/incident-response-rule.service';
import {WorkflowActionsService} from '../shared/services/workflow-actions.service';
import {IncidentRuleType} from '../shared/type/incident-rule.type';
import {minLengthArray} from "../../rule-management/custom-validators";

@Component({
  selector: 'app-playbook-builder',
  templateUrl: './playbook-builder.component.html',
  styleUrls: ['./playbook-builder.component.scss']
})
export class PlaybookBuilderComponent implements OnInit, OnDestroy {

  @Input() alert: any;
  @Input() rule: IncidentRuleType;
  @ViewChild('autocomplete') autocomplete: ElementRef;
  @Output() ruleCreated = new EventEmitter<boolean>();
  step = 1;
  stepCompleted: number[] = [];
  creating = false;
  formRule: FormGroup;
  agents: NetScanType[];
  platforms: string[];
  command = '';
  exist = true;
  typing = true;
  rulePrefix: string = createElementPrefix(PrefixElementEnum.INCIDENT_RESPONSE_AUTOMATION);
  maxLength = 512;
  viewportHeight: number;

  constructor(private incidentResponseRuleService: IncidentResponseRuleService,
              public activeModal: NgbActiveModal,
              private fb: FormBuilder,
              public inputClass: InputClassResolve,
              private utmToastService: UtmToastService,
              private route: ActivatedRoute,
              private workflowService: WorkflowActionsService,
              private router: Router) {

    this.formRule = this.fb.group({
      id: [null],
      name: ['', Validators.required],
      description: ['', Validators.required],
      conditions: this.fb.array([], minLengthArray(1)),
      command: ['', Validators.required],
      actions: [[]],
      active: [true],
      agentType: [false],
      excludedAgents: [[]],
      defaultAgent: [''],
      agentPlatform: ['', Validators.required]
    });
    this.viewportHeight = window.innerHeight;
  }

  @HostListener('window:resize', ['$event'])
  onResize(event: Event) {
    this.viewportHeight = window.innerHeight;
  }

  ngOnInit() {
    this.route.queryParams
      .pipe(
        filter(params => !!params && !!params.id),
        map(params => params.id),
        switchMap(id => {
          return this.incidentResponseRuleService.find(id)
            .pipe(
              map(res => res.body));
        })
      ).subscribe(rule => {

        const isTemplate = this.route.snapshot.queryParams.template === 'true';
        const ruleData = isTemplate ? { ...rule, id: null } : rule;
        const actions = isTemplate ? ruleData.actions.map(item => {
          return {
            ...item,
            id: null
          };
        }) : ruleData.actions;

        this.rule = ruleData;
        this.rule.actions = actions;

        this.exist = false;
        this.typing = false;
        this.rulePrefix = getElementPrefix(this.rule.name);


        this.formRule.patchValue(ruleData, { emitEvent: false });

        const name = this.formRule.get('name').value;
        this.formRule.get('name').setValue(this.replacePrefixInName(name));

        for (const condition of rule.conditions) {

          const ruleCondition = this.fb.group({
            field: [condition.field, Validators.required],
            value: [condition.value, Validators.required],
            operator: [condition.operator]
          });

          this.ruleConditions.push(ruleCondition);
        }
        this.command = this.rule.command;
        this.formRule.get('excludedAgents').setValue(this.rule.excludedAgents);
        this.formRule.get('agentType').setValue(this.rule.excludedAgents.length === 0 && this.rule.defaultAgent !== '');
        this.formRule.get('defaultAgent').setValue(this.rule.defaultAgent);
        this.rule.actions.forEach(action => this.workflowService.addActions(action));

      },
      error => {
        this.utmToastService.showError('Error', 'An error has occurred while fetching a rule');
      });

    this.route.queryParams
      .pipe(
        filter(params => !!params && !!params.alertName),
        map(params => params.alertName)).subscribe((alertName) => {
        this.addRuleCondition();
        const rc = this.ruleConditions.at(this.ruleConditions.length - 1);
        rc.patchValue({
          field: 'name',
          value: alertName,
        });
      });

    this.formRule.get('name').valueChanges.pipe(debounceTime(1000)).subscribe(value => {
      this.searchRule(this.rulePrefix + value);
    });
  }



  get ruleConditions() {
    return this.formRule.get('conditions') as FormArray;
  }


  replacePrefixInName(name: string) {
    return name.replace(this.rulePrefix, '');
  }

  addRuleCondition() {
    const ruleCondition = this.fb.group({
      field: ['', Validators.required],
      value: ['', Validators.required],
      operator: [ElasticOperatorsEnum.IS]
    });

    this.ruleConditions.push(ruleCondition);
  }

  getValueFromAlert(field: string) {
    return getValueFromPropertyPath(this.alert, field, null);
  }

  searchRule(rule: string) {
    this.typing = true;
    this.exist = true;
    setTimeout(() => {
      const req = {
        'name.contains': rule
      };
      this.incidentResponseRuleService.query(req).subscribe(response => {
        this.exist = response.body.length > 0;
        this.typing = false;
      });
    }, 1000);
  }

  getMenuHeight() {
    return 100 - ((150 / this.viewportHeight) * 100) + 'vh';
  }

  createRule() {
    if (this.rule.id) {
      this.editRule();
    } else {
      this.saveRule();
    }
  }

  saveRule() {
    const action = 'created';
    const actionError = 'creating';
    this.incidentResponseRuleService.create(this.formRule.value)
      .subscribe(() => {
            this.utmToastService.showSuccessBottom('Flow ' + action + ' successfully');
            this.router.navigate(['soar/flows']);
    }, () => this.errorSaving(actionError));
  }

  editRule() {
    const action = 'edited';
    const actionError = 'editing';
    this.formRule.get('command').setValue(this.command);
    this.incidentResponseRuleService.update(this.formRule.value).subscribe(() => {
      this.utmToastService.showSuccessBottom('Flow ' + action + ' successfully');
      this.router.navigate(['soar/flows']);
    }, () => this.errorSaving(actionError));
  }

  errorSaving(action: string) {
    const ruleName: string = this.formRule.get('name').value;
    this.formRule.get('name').setValue(this.replacePrefixInName(ruleName));
    this.utmToastService.showError('Error  ' + action + ' flow',
      'An error has occur while trying to ' + action + ' an flow, please contact support team');
  }

  ngOnDestroy(): void {
    this.workflowService.clear();
  }
}
