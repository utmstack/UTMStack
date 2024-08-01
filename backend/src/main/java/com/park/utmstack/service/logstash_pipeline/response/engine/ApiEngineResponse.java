package com.park.utmstack.service.logstash_pipeline.response.engine;

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

    public String getVersion() {
        return version;
    }

    public void setVersion(String version) {
        this.version = version;
    }

    public String getStatus() {
        return status;
    }

    public void setStatus(String status) {
        this.status = status;
    }

}
