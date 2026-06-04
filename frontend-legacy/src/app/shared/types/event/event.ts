export interface Geolocation {
  country: string;
  city: string;
  latitude: number;
  longitude: number;
  asn: number;
  aso: string;
  countryCode: string;
  accuracy: number;
}

export interface Side {
  bytesSent: number;
  bytesReceived: number;
  packagesSent: number;
  packagesReceived: number;
  connections: number;
  usedCpuPercent: number;
  usedMemPercent: number;
  totalCpuUnits: number;
  totalMem: number;
  ip: string;
  host: string;
  user: string;
  group: string;
  port: number;
  domain: string;
  fqdn: string;
  mac: string;
  process: string;
  geolocation: Geolocation;
  file: string;
  path: string;
  hash: string;
  url: string;
  email: string;
}

export interface Event {
  id: string;
  timestamp: string;
  deviceTime: string;
  dataType: string;
  dataSource: string;
  tenantId: string;
  tenantName: string;
  raw: string;
  log: { [key: string]: any };
  target: Side;
  origin: Side;
  protocol: string;
  connectionStatus: string;
  statusCode: number;
  actionResult: string;
  action: string;
  command: string;
  severity: string;
}
