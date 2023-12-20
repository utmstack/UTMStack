import {Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';
import {ComplianceStandardSectionType} from '../type/compliance-standard-section.type';

@Injectable({providedIn: 'root'})
export class CpStandardSectionBehavior {
  $standardSection: BehaviorSubject<ComplianceStandardSectionType> = new BehaviorSubject<ComplianceStandardSectionType>(null);
  $updateStandardSection: BehaviorSubject<boolean> = new BehaviorSubject<boolean>(null);
}
