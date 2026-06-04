import {Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';
import {AssetFilterType} from '../types/asset-filter.type';

@Injectable({
  providedIn: 'root'
})
export class AssetGroupFiltersBehavior {
  $assetGroupFilter = new BehaviorSubject<AssetFilterType>(null);
  $assetGroupAppliedFilter = new BehaviorSubject<AssetFilterType>(null);
}
