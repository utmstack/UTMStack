import {Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';

export enum VersionType {
  COMMUNITY = 'COMMUNITY',
  ENTERPRISE = 'ENTERPRISE',
}

@Injectable({
  providedIn: 'root'
})
export class VersionTypeService {
 private versionTypeBehavior = new BehaviorSubject<VersionType>(VersionType.COMMUNITY);
 versionType$ = this.versionTypeBehavior.asObservable();

 changeVersionType(versionType: VersionType) {
   this.versionTypeBehavior.next(versionType);
 }

 versionType(): VersionType {
   return this.versionTypeBehavior.getValue();
 }
}
