package com.park.utmstack.service.logstash_pipeline.response.pipeline;

import com.fasterxml.jackson.annotation.JsonIgnore;

public class PipelineLastError {
    String message;
    @JsonIgnore
    String [] backtrace;

    public PipelineLastError(){}
    public PipelineLastError(String message, String[] backtrace) {
        this.message = message;
        this.backtrace = backtrace;
    }

    public String getMessage() {
        return message;
    }

    public void setMessage(String message) {
        this.message = message;
    }

    public String[] getBacktrace() {
        return backtrace;
    }

    public void setBacktrace(String[] backtrace) {
        this.backtrace = backtrace;
    }
}
