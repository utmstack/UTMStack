export interface NavAction {
  readonly label: string;
  readonly iconClass: string;
  readonly route: string;
}

export const FEDERATION_NAV_ACTIONS: ReadonlyArray<NavAction> = [
  {label: 'Log Explorer', iconClass: 'icon-search4', route: '/discover/log-analyzer'},
  {label: 'Alert Management', iconClass: 'icon-bell2', route: '/data/alert/view'},
  {label: 'SOAR Flows', iconClass: 'icon-stack-play', route: '/soar/flows'},
  {label: 'Data Sources', iconClass: 'icon-database', route: '/data-sources'}
];
