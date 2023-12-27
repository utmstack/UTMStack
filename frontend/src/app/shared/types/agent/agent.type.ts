export class AgentType {
  ip: string;
  hostname: string;
  os: string;
  status: AgentStatusEnum;
  platform: string;
  version: string;
  agentKey: string;
  id: number;
  lastSeen: Date;
  mac: string;
  osMajorVersion: string;
  osMinorVersion: string;
  aliases: string;
  addresses: string;
}

export class AgentCommandType {
  createdAt: string;
  updatedAt: string;
  agentId: number;
  command: string;
  commandStatus: string;
  result: string;
  executedBy: string;
  cmdId: string;
  agent: AgentType;
  reason: string;
  originType: string;
  originId: string;
}

export class AgentCommandDateType {
  seconds: string;
  nanos: number;
}


export enum AgentStatusEnum {
  ONLINE = 'ONLINE',
  OFFLINE = 'OFFLINE'
}
