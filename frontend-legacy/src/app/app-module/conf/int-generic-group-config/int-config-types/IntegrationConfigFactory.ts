import { Injectable } from '@angular/core';
import {GroupTypeEnum} from '../../../shared/enum/group-type.enum';
import { CollectorConfiguration } from './collector-configuration';
import {GenericConfiguration} from './generic-configuration';
import { IntegrationConfig } from './integration-config';

@Injectable()
export class IntegrationConfigFactory {
    constructor(
        private collectorConfig: CollectorConfiguration,
        private genericConfig: GenericConfiguration
    ) {}

    getConfiguration(type: GroupTypeEnum): IntegrationConfig {
        switch (type) {
            case GroupTypeEnum.COLLECTOR:
                return this.collectorConfig;
            case GroupTypeEnum.TENANT:
                return this.genericConfig;
            default:
                throw new Error('Unknown configuration type');
        }
    }
}
