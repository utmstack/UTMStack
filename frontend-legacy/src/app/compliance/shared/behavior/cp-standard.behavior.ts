import {Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';
import {ComplianceStandardType} from '../type/compliance-standard.type';

@Injectable({providedIn: 'root'})
export class CpStandardBehavior {
  $standard: BehaviorSubject<ComplianceStandardType> = new BehaviorSubject<ComplianceStandardType>(null);
  $updateStandard: BehaviorSubject<boolean> = new BehaviorSubject<boolean>(null);
}
