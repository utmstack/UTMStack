package com.park.utmstack.service.logstash_pipeline.response.jvm_stats;

public class LogstashApiJvmResponse {
    String version;
    String status;

    public LogstashApiJvmResponse(){
        this.version = "2.0.0";
        this.status = "down";
    }
    public LogstashApiJvmResponse(String host, String version,
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
