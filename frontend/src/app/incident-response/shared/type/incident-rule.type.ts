import {ElasticFilterType} from '../../../shared/types/filter/elastic-filter.type';
import {IncidentResponseActionTemplate} from "../services/incident-response-action-template.service";

export class IncidentRuleType {
  id: number;
  name: string;
  description: string;
  conditions: ElasticFilterType[];
  command: string;
  active: boolean;
  excludedAgents: string;
  agentPlatform: string;
  shell?: string;
  createdBy: string;
  createdDate: Date;
  lastModifiedBy: string;
  lastModifiedDate: Date;
  defaultAgent: string;
  actions?: IncidentResponseActionTemplate[];
}


export class IncidentRuleSelectType {
  agentPlatform: string[];
  users: string[];
}

