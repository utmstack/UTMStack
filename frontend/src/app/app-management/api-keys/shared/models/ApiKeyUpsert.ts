export interface ApiKeyUpsert {
  name: string;
  allowedIp?: string[];
  expiresAt?: Date;
}
