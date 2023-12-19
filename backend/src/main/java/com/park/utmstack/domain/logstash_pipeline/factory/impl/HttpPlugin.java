package com.park.utmstack.domain.logstash_pipeline.factory.impl;

import com.park.utmstack.domain.logstash_pipeline.enums.InputConfigTypes;
import com.park.utmstack.domain.logstash_pipeline.enums.InputPlugin;
import com.park.utmstack.domain.logstash_pipeline.enums.ValidationRegexp;
import com.park.utmstack.domain.logstash_pipeline.factory.IPluginInput;
import com.park.utmstack.domain.logstash_pipeline.types.InputConfiguration;
import com.park.utmstack.domain.logstash_pipeline.types.InputConfigurationKey;
import com.park.utmstack.service.logstash_pipeline.UtmLogstashInputConfigurationService;
import org.springframework.stereotype.Component;

import java.util.ArrayList;
import java.util.List;

@Component
public class HttpPlugin implements IPluginInput {
    Boolean isSsl;
    private final UtmLogstashInputConfigurationService utmLogstashInputConfigurationService;
    public HttpPlugin(UtmLogstashInputConfigurationService utmLogstashInputConfigurationService){
        this.utmLogstashInputConfigurationService = utmLogstashInputConfigurationService;
    }
    public HttpPlugin withSsl(Boolean isSsl){
        this.isSsl = isSsl;
        return this;
    }

    @Override
    public InputConfiguration getConfiguration() throws Exception {
        InputConfiguration tcpConfig = new InputConfiguration();
        List<InputConfigurationKey> configs = new ArrayList<>();
        String http = InputPlugin.HTTP.name();
        List<String> availablePorts = utmLogstashInputConfigurationService
            .getAvailableConfigsByType(http.toLowerCase());
        if(!availablePorts.isEmpty()){
            configs.add(InputConfigurationKey.builder()
                .withConfKey(getKeyBySsl())
                .withConfValue(availablePorts.get(0))
                .withConfType(InputConfigTypes.PORT.getValue())
                .withConfRequired(true)
                .withConfValidationRegex(ValidationRegexp.PORT.getRegexp())
                .withSystemOwner(false)
                .build());
            tcpConfig.setConfigs(configs);
            tcpConfig.setInputWithSsl(this.isSsl);
            tcpConfig.setInputPlugin(http.toLowerCase());
            tcpConfig.setInputPrettyName(getPrettyNameBySsl(http));
            tcpConfig.setSystemOwner(false);
            return tcpConfig;
        }
        throw new Exception("Sorry, no ports available");
    }
    public String getKeyBySsl(){
        return this.isSsl?"http_ssl_port":"http_port";
    }
    public String getPrettyNameBySsl(String proto){
        return this.isSsl?proto+" with SSL":proto;
    }
}
