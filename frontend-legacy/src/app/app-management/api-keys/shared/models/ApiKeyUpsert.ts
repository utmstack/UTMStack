export interface ApiKeyUpsert {
  id: string;
  name: string;
  allowedIp?: string[];
  expiresAt?: Date;
}
