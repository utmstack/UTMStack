import {ElasticFilterType} from '../../../shared/types/filter/elastic-filter.type';

export class IncidentRuleType {
  id: number;
  name: string;
  description: string;
  conditions: ElasticFilterType[];
  command: string;
  active: boolean;
  excludedAgents: string;
  agentPlatform: string;
  createdBy: string;
  createdDate: Date;
  lastModifiedBy: string;
  lastModifiedDate: Date;
  defaultAgent: string;
}


export class IncidentRuleSelectType {
  agentPlatform: string[];
  users: string[];
}

