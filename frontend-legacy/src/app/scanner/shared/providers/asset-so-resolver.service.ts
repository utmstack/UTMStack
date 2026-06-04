import {Injectable} from '@angular/core';
import {AssetModel} from '../model/assets/asset.model';

@Injectable({
  providedIn: 'root'
})
export class AssetSoResolverService {

  constructor() {
  }


  public resolveAssetSoName(asset: AssetModel): string {
    if (asset.host.detail !== null) {
      const index = asset.host.detail.findIndex(value => value.name === 'best_os_txt');
      if (index !== -1) {
        return asset.host.detail[index].value;
      }
    }
    return 'Unknown';
  }

  public resolveAssetSoIcon(asset: AssetModel): string {
    if (asset.host.detail !== null) {
      const index = asset.host.detail.findIndex(value => value.name === 'best_os_txt');
      if (index !== -1) {
        if (asset.host.detail[index].value.toLowerCase().includes('windows')) {
          return 'icon-windows8';
        } else if (asset.host.detail[index].value.toLowerCase().includes('linux')) {
          return 'icon-tux';
        } else if (asset.host.detail[index].value.toLowerCase().includes('mac')) {
          return 'icon-apple2';
        } else if (asset.host.detail[index].value.toLowerCase().includes('possible conflict')) {
          return 'icon-power2';
        }
      }
    }
    return 'icon-question3';
  }

  public resolveAssetSoClass(asset: AssetModel): string {
    if (asset.host.detail !== null) {
      const index = asset.host.detail.findIndex(value => value.name === 'best_os_txt');
      if (index !== -1) {
        if (asset.host.detail[index].value.toLowerCase().includes('windows')) {
          // border-indigo-400
          return 'text-indigo-400';
        } else if (asset.host.detail[index].value.toLowerCase().includes('linux')) {
          // border-warning-400
          return 'text-warning-400';
        } else if (asset.host.detail[index].value.toLowerCase().includes('mac')) {
          // border-slate-400
          return 'text-slate-400';
        } else if (asset.host.detail[index].value.toLowerCase().includes('possible conflict')) {
          // border-orange-400
          return 'text-orange-400';
        }
      }
    }
    // border-purple-400
    return 'text-purple-400';
  }

}
