export interface NavAction {
  readonly label: string;
  readonly iconClass: string;
  readonly route: string;
  readonly counter: boolean
}

export const FEDERATION_NAV_ACTIONS: ReadonlyArray<NavAction> = [
  {label: 'Alert Management', iconClass: 'PROMOTION.svg', route: '/data/alert/view', counter:true},
  {label: 'Log Explorer', iconClass: 'ANALYTICS.svg', route: '/discover/log-analyzer',counter:false},
  {label: 'SOAR Flows', iconClass: 'VISION.svg', route: '/soar/flows',counter:false},
  {label: 'Data Sources', iconClass: 'NETWORK.svg', route: '/data-sources',counter:false}
];
