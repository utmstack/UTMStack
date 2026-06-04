import {AfterViewInit, Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {Router} from '@angular/router';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {Observable, of} from 'rxjs';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {InputClassResolve} from '../../../../shared/util/input-class-resolve';
import {IncidentResponseRuleService} from '../../services/incident-response-rule.service';
import {PlaybookService} from '../../services/playbook.service';
import {IncidentActionType} from '../../type/incident-action.type';

@Component({
  selector: 'app-incident-response-command-create',
  templateUrl: './new-playbook.component.html',
  styleUrls: ['./new-playbook.component.scss'],
  providers: [PlaybookService]
})
export class NewPlaybookComponent implements OnInit, AfterViewInit {
  @Input() action: IncidentActionType;
  @Output() actionCreated = new EventEmitter<IncidentActionType>();

  request = {
    page: 0,
    size: 25,
    sort: '',
    'active.equals': null,
    'agentPlatform.equals': null,
    'createdBy.equals': null,
    'systemOwner.equals': true
  };
  platforms: string[];

  platforms$: Observable<string[]>;
  loadingPlatform = false;

  constructor(public activeModal: NgbActiveModal,
              public inputClass: InputClassResolve,
              private utmToastService: UtmToastService,
              private incidentResponseRuleService: IncidentResponseRuleService,
              public playbookService: PlaybookService,
              private router: Router) {
  }

  ngOnInit() {}

  ngAfterViewInit(): void {
    this.playbookService.loadData({...this.request});
  }

  createNewPlaybook(params?: any) {
    this.router.navigate(['/soar/create-flow'], { queryParams: params });
    this.activeModal.close();
  }

  searchByRule($event: string | number) {
    this.request['name.contains'] = $event;
    this.playbookService.loadData({
      ...this.request,
      page: 0,
    });
  }

  trackByFn(index: number, item: any) {
    return item.id || index;
  }
}
