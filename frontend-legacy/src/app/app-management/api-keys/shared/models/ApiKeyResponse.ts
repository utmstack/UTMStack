export interface ApiKeyResponse {
  id: string;
  name: string;
  allowedIp: string[];
  createdAt: string;
  expiresAt?: string;
  generatedAt?: string;
}
