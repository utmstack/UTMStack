// Class to enum data nature
export enum NatureDataPrefixEnum {
  CORRELATION = 'correlation.',
  GLOBAL = 'global.',
  VULNERABILITY = 'vulnerability.',
  TIMESTAMP = '@timestamp',
  ALERT = 'alert.',
  EVENT = 'logx.'
}

export enum DataNatureTypeEnum {
  ALERT = 'v11-alert-*',
  EVENT = 'v11-log-*',
  VULNERABILITY = 'VULNERABILITY'
}
