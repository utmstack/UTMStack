package com.park.utmstack.domain;


import javax.persistence.Column;
import javax.persistence.Entity;
import javax.persistence.Id;
import javax.persistence.Table;
import java.io.Serializable;

/**
 * A UtmAlertSocaiProcessingRequest.
 */
@Entity
@Table(name = "utm_alert_socai_processing_request")
public class UtmAlertSocaiProcessingRequest implements Serializable {

    private static final long serialVersionUID = 1L;

    public UtmAlertSocaiProcessingRequest(String alertId) {
        this.alertId = alertId;
    }

    public UtmAlertSocaiProcessingRequest() {
    }

    @Id
    @Column(name = "alert_id")
    private String alertId;

    public String getAlertId() {
        return alertId;
    }

    public void setAlertId(String alertId) {
        this.alertId = alertId;
    }
}
