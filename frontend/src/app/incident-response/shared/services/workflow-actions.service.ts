import {Injectable} from '@angular/core';
import {BehaviorSubject, Observable} from 'rxjs';
import {map} from 'rxjs/operators';
import {ActionConditionalEnum} from '../component/action-conditional/action-conditional.component';
import {IncidentResponseActionTemplate} from './incident-response-action-template.service';

@Injectable({
  providedIn: 'root'
})
export class WorkflowActionsService {

  private actionsBehaviorSubject: BehaviorSubject<IncidentResponseActionTemplate[]> = new BehaviorSubject([]);
  actions$ = this.actionsBehaviorSubject.asObservable();

  readonly command$: Observable<string> = this.actions$.pipe(
    map(actions => this.buildCommand(actions))
  );

  buildCommand(actions: IncidentResponseActionTemplate[] = this.getActions()): string {
    if (!actions || actions.length === 0) {
      return '';
    }

    if (actions.length === 1) {
      return actions[0].command;
    }

    return actions.map((action, index) => {
      const operator = index === 0 ? ''
        : action.conditional.key === ActionConditionalEnum.SUCCESS ? '&&'
          : action.conditional.key === ActionConditionalEnum.FAILURE ? '||'
            : ';';

      return `${operator} ${action.command}`.trim();
    }).join(' ').trim();
  }

  inferConditionals(command: string, actions: IncidentResponseActionTemplate[]): IncidentResponseActionTemplate[] {
    if (!actions || actions.length === 0) {
      return actions || [];
    }

    if (!command || actions.length === 1) {
      return [{ ...actions[0], conditional: { key: ActionConditionalEnum.ALWAYS, value: ';' } }];
    }

    const result: IncidentResponseActionTemplate[] = [];
    let cursor = 0;

    actions.forEach((action, index) => {
      const idx = command.indexOf(action.command, cursor);

      if (index === 0 || idx === -1) {
        result.push({ ...action, conditional: { key: ActionConditionalEnum.ALWAYS, value: ';' } });
      } else {
        const gap = command.slice(cursor, idx).trim();
        const conditional = gap === '&&' ? { key: ActionConditionalEnum.SUCCESS, value: '&&' }
          : gap === '||' ? { key: ActionConditionalEnum.FAILURE, value: '||' }
            : { key: ActionConditionalEnum.ALWAYS, value: ';' };
        result.push({ ...action, conditional });
      }

      if (idx !== -1) {
        cursor = idx + action.command.length;
      }
    });

    return result;
  }

  addActions(action: any) {
    const actions = this.actionsBehaviorSubject.value ? this.actionsBehaviorSubject.value : [];

    this.actionsBehaviorSubject.next([...actions, {
      ...action,
      conditional: action.conditional ? action.conditional : { key: ActionConditionalEnum.ALWAYS, value: ';'},
    }]);
  }

  updateAction(index: number, action: any) {
    const actions = this.getActions();
    if (index < 0 || index >= actions.length) {
      return;
    }

    const newActions = [...actions];
    newActions[index] = { ...action };

    this.actionsBehaviorSubject.next(newActions);
  }

  deleteAction(action: any) {
    const actions = this.actionsBehaviorSubject.value ? this.actionsBehaviorSubject.value : [];
    this.actionsBehaviorSubject.next(actions.filter(act => act !== action));
  }

  clear() {
    this.actionsBehaviorSubject.next([]);
  }

  getActions() {
    return this.actionsBehaviorSubject.value ? this.actionsBehaviorSubject.value : [];
  }
}
