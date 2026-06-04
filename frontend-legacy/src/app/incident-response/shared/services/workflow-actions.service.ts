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
    map(actions => {
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
    })
  );

  addActions(action: any) {
    const actions = this.actionsBehaviorSubject.value ? this.actionsBehaviorSubject.value : [];

    this.actionsBehaviorSubject.next([...actions, {
      ...action,
      conditional: action.conditional ? action.conditional : { key: ActionConditionalEnum.ALWAYS, value: ';'},
    }]);
  }

  updateAction(action: any) {
    const actions = this.actionsBehaviorSubject.value ? this.actionsBehaviorSubject.value : [];

    const index = actions.findIndex((act: any) => act.id === action.id);

    const newActions = [...actions];
    newActions[index] = {
      ...action,
    };

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
