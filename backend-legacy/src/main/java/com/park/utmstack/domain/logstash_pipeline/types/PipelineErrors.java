package com.park.utmstack.domain.logstash_pipeline.types;

import java.util.List;

public class PipelineErrors {
    List<Validation> validationErrors;

    public PipelineErrors(List<Validation> validationErrors) {
        this.validationErrors = validationErrors;
    }

    public List<Validation> getValidationErrors() {
        return validationErrors;
    }

    public void setValidationErrors(List<Validation> validationErrors) {
        this.validationErrors = validationErrors;
    }
}
