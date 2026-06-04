package com.park.utmstack.service.logstash_pipeline.response.engine;

import lombok.Getter;
import lombok.Setter;

@Setter
@Getter
public class ApiEngineResponse {
    String version;
    String status;

    public ApiEngineResponse(){
        this.version = "2.0.0";
        this.status = "down";
    }
    public ApiEngineResponse(String host, String version,
                             String status) {
        this.version = version;
        this.status = status;
    }

}
