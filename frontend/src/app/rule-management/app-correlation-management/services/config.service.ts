
import {Injectable} from '@angular/core';
import {BehaviorSubject, Observable} from 'rxjs';
import { Actions } from '../models/config.type';


@Injectable()
export class ConfigService {

    private actionBehaviorSubject = new BehaviorSubject<string>(null);
    action$ = this.actionBehaviorSubject.asObservable();

    constructor() {
    }

    onAction(action: Actions){
        this.actionBehaviorSubject.next(action);
    }
}
