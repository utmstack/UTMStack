import {UtmDashboardType} from '../../../../shared/chart/types/dashboard/utm-dashboard.type';

export function buildDashboardUrl(dashboard: UtmDashboardType): string {
  let str = dashboard.name;
  str = str.replace(/\W+(?!$)/g, '-').toLowerCase();
  str = str.replace(/\W$/, '').toLowerCase();
  return 'dashboard/render/' + dashboard.id +
    '/' + str;
}
