package com.park.utmstack.domain.application_events.types;

import com.fasterxml.jackson.annotation.JsonProperty;

public class ApplicationEvent {
    @JsonProperty("@timestamp")
    private String timestamp;
    private String source;
    private String message;
    private String type;

    public ApplicationEvent() {
    }

    public ApplicationEvent(String timestamp, String source, String message, String type) {
        this.timestamp = timestamp;
        this.source = source;
        this.message = message;
        this.type = type;
    }

    public String getTimestamp() {
        return timestamp;
    }

    public void setTimestamp(String timestamp) {
        this.timestamp = timestamp;
    }

    public String getSource() {
        return source;
    }

    public void setSource(String source) {
        this.source = source;
    }

    public String getMessage() {
        return message;
    }

    public void setMessage(String message) {
        this.message = message;
    }

    public String getType() {
        return type;
    }

    public void setType(String type) {
        this.type = type;
    }

    public static Builder builder() {
        return new Builder();
    }

    public static class Builder {
        private String timestamp;
        private String source;
        private String message;
        private String type;

        public ApplicationEvent build() {
            return new ApplicationEvent(timestamp, source, message, type);
        }

        public Builder timestamp(String timestamp) {
            this.timestamp = timestamp;
            return this;
        }

        public Builder source(String source) {
            this.source = source;
            return this;
        }

        public Builder message(String message) {
            this.message = message;
            return this;
        }

        public Builder type(String type) {
            this.type = type;
            return this;
        }
    }
}
