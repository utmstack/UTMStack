export interface NavAction {
  readonly label: string;
  readonly iconClass: string;
  readonly route: string;
}

export const FEDERATION_NAV_ACTIONS: ReadonlyArray<NavAction> = [
  {label: 'Log Explorer', iconClass: 'ANALYTICS.svg', route: '/discover/log-analyzer'},
  {label: 'Alert Management', iconClass: 'PROMOTION.svg', route: '/data/alert/view'},
  {label: 'SOAR Flows', iconClass: 'VISION.svg', route: '/soar/flows'},
  {label: 'Data Sources', iconClass: 'NETWORK.svg', route: '/data-sources'}
];
