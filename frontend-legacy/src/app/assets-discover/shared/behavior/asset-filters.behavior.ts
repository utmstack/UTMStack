import {Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';
import {AssetFilterType} from '../types/asset-filter.type';

@Injectable({
  providedIn: 'root'
})
export class AssetFiltersBehavior {
  $assetFilter = new BehaviorSubject<AssetFilterType>(null);
  $assetAppliedFilter = new BehaviorSubject<AssetFilterType>(null);
}
