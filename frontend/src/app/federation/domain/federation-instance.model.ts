export interface FederationInstance {
  id: number;
  name: string;
  baseUrl: string;
  tlsSkipVerify: boolean;
  createdAt: string;
  updatedAt: string;
  version?: string;
}
