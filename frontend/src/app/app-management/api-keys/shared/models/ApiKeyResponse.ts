export interface ApiKeyResponse {
  id: string;
  name: string;
  allowedIp: string[];
  createdAt: Date;
  expiresAt?: Date;
  generatedAt?: Date;
}
