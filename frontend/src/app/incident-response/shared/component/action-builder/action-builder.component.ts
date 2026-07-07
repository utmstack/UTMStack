import {Component, Input, OnDestroy, OnInit} from '@angular/core';
import {FormGroup} from '@angular/forms';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {Observable, of, Subject} from 'rxjs';
import {catchError, filter, finalize, map, takeUntil, tap} from 'rxjs/operators';
import {UtmNetScanService} from '../../../../assets-discover/shared/services/utm-net-scan.service';
import {NetScanType} from '../../../../assets-discover/shared/types/net-scan.type';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {InputClassResolve} from '../../../../shared/util/input-class-resolve';
import {WorkflowActionsService} from '../../services/workflow-actions.service';
import {ActionConditionalEnum} from '../action-conditional/action-conditional.component';
import {ActionTerminalComponent} from '../action-terminal/action-terminal.component';

@Component({
  selector: 'app-action-builder',
  templateUrl: './action-builder.component.html',
  styleUrls: ['./action-builder.component.scss']
})
export class ActionBuilderComponent implements OnInit, OnDestroy {

  @Input() group: FormGroup;
  agents: any[];

  platforms$: Observable<string[]>;
  loadingPlatforms = false;
  agents$: Observable<NetScanType[]>;
  loadingAgents = false;
  noPlatforms = false;

  workflow$: Observable<any[]>;
  command$: Observable<string>;
  destroy$: Subject<void> = new Subject<void>();

  constructor(private utmNetScanService: UtmNetScanService,
              public inputClass: InputClassResolve,
              private utmToastService: UtmToastService,
              public workflowActionsService: WorkflowActionsService,
              private modalService: NgbModal) { }

  ngOnInit() {
    this.platforms$ = this.getPlatforms();

    this.workflow$ = this.workflowActionsService.actions$
      .pipe(takeUntil(this.destroy$));

    this.command$ = this.workflowActionsService.command$
      .pipe(takeUntil(this.destroy$),
            tap((command) => {
              this.group.get('actions').setValue(this.workflowActionsService.getActions());
              this.group.get('command').setValue(command);
            })
        );

    this.group.get('agentPlatform').valueChanges
      .pipe(
        takeUntil(this.destroy$),
        filter(val => !!val)
      )
      .subscribe(platform => {
        this.loadingAgents = true;
        this.agents$ = this.fetchAgents(platform);
      });
  }

  getPlatforms() {
    this.loadingPlatforms = true;
    return this.utmNetScanService.getAssetsPlatforms().pipe(
      map(res => res.body || []),
      tap(platforms => {
        this.noPlatforms = platforms.length === 0;
      }),
      catchError(err => {
        this.utmToastService.showError('Error fetching', 'An error has occurred while fetching platforms');
        return of([]);
      }),
      finalize(() => {
        this.loadingPlatforms = false;
      })
    );
  }

  getAgents(platform: any) {
    this.group.get('excludedAgents').reset();
    this.group.get('defaultAgent').reset();

    this.loadingAgents = true;
    this.agents$ = this.fetchAgents(platform);
  }

  onChangeToggle($event) {
    if ($event) {
      this.group.get('excludedAgents').setValue([]);
    } else {
      this.group.get('defaultAgent').setValue('');
    }
    this.group.get('agentType').setValue($event);
  }

  onChangeExclude($event: any) {
    const agentType = this.group.get('agentType').value;
    if (!agentType) {
      const hostnames = $event.map((value: { assetName: any; }) => value.assetName);
      this.group.get('excludedAgents').setValue(hostnames);
    }
  }

  fetchAgents(platform: string) {
    return this.utmNetScanService.query({page: 0, size: 10000, agent: true, osPlatform: platform})
      .pipe(
        map(res => res.body || []),
        tap(agents => {
          if (agents.length === 1) {
            this.group.get('excludedAgents').disable();
          } else {
            this.group.get('excludedAgents').enable();
          }
        }),
        catchError(err => {
          this.utmToastService.showError('Error fetching', 'An error has occurred while fetching agents');
          return of([]);
        }),
        finalize(() => {
          this.loadingAgents = false;
        })
      );
  }

  openActionSidebar(action: any = null) {
    const dialogRef = this.modalService.open(ActionTerminalComponent, {size: 'lg', centered: true});
    dialogRef.componentInstance.action = action || null;

    dialogRef.result.then(
      result => {
        if (result) {
          if (action) {
            this.removeAction(action);
          }
          this.workflowActionsService.addActions({
            ...action,
            title: result.title,
            description: result.description,
            command: result.command,
          });
        }
      },
      reason => {
        console.log('Modal cerrado por:', reason);
      }
    );
  }

  removeAction(action: any) {
    this.workflowActionsService.deleteAction(action);
  }

  select(always: string) {

  }

  updateAction(action: any, index: number, $event: { key: ActionConditionalEnum; value: string }) {
    this.workflowActionsService.updateAction(index, {
      ...action,
      conditional: $event
    });
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }
}
