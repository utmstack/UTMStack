import {Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';
import {AssetFieldFilterEnum} from '../enums/asset-field-filter.enum';
import {CollectorFieldFilterEnum} from "../enums/collector-field-filter.enum";

@Injectable({
  providedIn: 'root'
})
/**
 * Use this behavior when you update the type of an asset, to trigger observable and refresh filter values
 */
export class AssetReloadFilterBehavior {
  $assetReloadFilter = new BehaviorSubject<AssetFieldFilterEnum | CollectorFieldFilterEnum>(null);
}
