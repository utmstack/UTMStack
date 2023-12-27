package com.park.utmstack.domain.logstash_pipeline.factory;

import com.park.utmstack.domain.logstash_pipeline.types.InputConfiguration;

public interface IPluginInput {
    InputConfiguration getConfiguration() throws Exception;
}
