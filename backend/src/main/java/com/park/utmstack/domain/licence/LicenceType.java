package com.park.utmstack.domain.licence;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;

import java.time.Instant;

@JsonInclude(value = JsonInclude.Include.NON_EMPTY)
public class LicenceType {

    private Details details;

    private Boolean valid;

    private Instant expire;

    public Details getDetails() {
        return details;
    }

    public void setDetails(Details details) {
        this.details = details;
    }

    public Boolean getValid() {
        return valid;
    }

    public void setValid(Boolean valid) {
        this.valid = valid;
    }

    public Instant getExpire() {
        return expire;
    }

    public void setExpire(Instant expire) {
        this.expire = expire;
    }

    public static class Details {

        private Instant activation;

        @JsonProperty("customer_email")
        private String customerEmail;

        @JsonProperty("customer_name")
        private String customerName;

        private Integer days;

        private String key;

        private Integer status;

        public Instant getActivation() {
            return activation;
        }

        public void setActivation(Instant activation) {
            this.activation = activation;
        }

        public String getCustomerEmail() {
            return customerEmail;
        }

        public void setCustomerEmail(String customerEmail) {
            this.customerEmail = customerEmail;
        }

        public String getCustomerName() {
            return customerName;
        }

        public void setCustomerName(String customerName) {
            this.customerName = customerName;
        }

        public Integer getDays() {
            return days;
        }

        public void setDays(Integer days) {
            this.days = days;
        }

        public String getKey() {
            return key;
        }

        public void setKey(String key) {
            this.key = key;
        }

        public Integer getStatus() {
            return status;
        }

        public void setStatus(Integer status) {
            this.status = status;
        }
    }

    public enum Status {
        INACTIVE(0),
        VALID(1);

        private final int status;

        public int status() {
            return status;
        }

        public static Status status(int status) {
            switch (status) {
                case 0:
                    return INACTIVE;
                case 1:
                    return VALID;
            }
            return null;
        }

        Status(int status) {
            this.status = status;
        }


    }
}
