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
public class TcpPlugin implements IPluginInput {
    Boolean isSsl;
    private final UtmLogstashInputConfigurationService utmLogstashInputConfigurationService;
    public TcpPlugin(UtmLogstashInputConfigurationService utmLogstashInputConfigurationService){
        this.utmLogstashInputConfigurationService = utmLogstashInputConfigurationService;
    }
    public TcpPlugin withSsl(Boolean isSsl){
        this.isSsl = isSsl;
        return this;
    }

    @Override
    public InputConfiguration getConfiguration() throws Exception {
        InputConfiguration tcpConfig = new InputConfiguration();
        List<InputConfigurationKey> configs = new ArrayList<>();
        String tcp = InputPlugin.TCP.name();
        List<String> availablePorts = utmLogstashInputConfigurationService
            .getAvailableConfigsByType(tcp.toLowerCase());
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
            tcpConfig.setInputPlugin(tcp.toLowerCase());
            tcpConfig.setInputPrettyName(getPrettyNameBySsl(tcp));
            tcpConfig.setSystemOwner(false);
            return tcpConfig;
        }
        throw new Exception("Sorry, no ports available");
    }
    public String getKeyBySsl(){
        return this.isSsl?"tcp_ssl_port":"tcp_port";
    }
    public String getPrettyNameBySsl(String proto){
        return this.isSsl?proto+" with SSL":proto;
    }
}
