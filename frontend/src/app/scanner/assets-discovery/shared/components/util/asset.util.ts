import {AssetModel} from '../../../../shared/model/assets/asset.model';

export function resolveAssetSoName(asset: AssetModel): string {
  const index = asset.host.detail.findIndex(value => value.name === 'best_os_txt');
  return asset.host.detail[index].value;
}

export function resolveAssetSoIcon(asset: AssetModel): string {
  const index = asset.host.detail.findIndex(value => value.name === 'best_os_txt');
  if (asset.host.detail[index].value.toLowerCase().includes('windows')) {
    return 'icon-windows8';
  } else if (this.asset.host.detail[index].value.toLowerCase().includes('linux')) {
    return 'icon-tux';
  } else if (this.asset.host.detail[index].value.toLowerCase().includes('mac')) {
    return 'icon-apple2';
  }
}

export function resolveAssetSoClass(asset: AssetModel): string {
  const index = asset.host.detail.findIndex(value => value.name === 'best_os_txt');
  if (asset.host.detail[index].value.toLowerCase().includes('windows')) {
    return 'border-indigo-400 text-indigo-400';
  } else if (this.asset.host.detail[index].value.toLowerCase().includes('linux')) {
    return 'border-warning-400 text-warning-400';
  } else if (this.asset.host.detail[index].value.toLowerCase().includes('mac')) {
    return 'border-slate-400 text-slate-400';
  }
}

export function resolveSeverityFixed(asset: AssetModel): string {
  return this.asset.host.severity.value.replace('.0', '');
}
