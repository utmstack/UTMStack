package com.park.utmstack.domain;


import org.hibernate.annotations.GenericGenerator;

import javax.persistence.*;
import java.io.Serializable;
import java.time.Instant;
import java.time.ZoneOffset;
import java.util.Objects;

/**
 * A UtmAlertLast.
 */
@Entity
@Table(name = "utm_alert_last")
public class UtmAlertLast implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GenericGenerator(name = "CustomIdentityGenerator", strategy = "com.park.utmstack.util.CustomIdentityGenerator")
    @GeneratedValue(generator = "CustomIdentityGenerator")
    private Long id;

    @Column(name = "last_alert_timestamp")
    private Instant lastAlertTimestamp;

    public UtmAlertLast() {
    }

    public UtmAlertLast(Instant lastAlertTimestamp) {
        this.id = 1L;
        this.lastAlertTimestamp = lastAlertTimestamp;
    }

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public Instant getLastAlertTimestamp() {
        return !Objects.isNull(lastAlertTimestamp) ? lastAlertTimestamp : Instant.now().atZone(ZoneOffset.UTC).toInstant();
    }

    public void setLastAlertTimestamp(Instant lastAlertTimestamp) {
        this.lastAlertTimestamp = lastAlertTimestamp;
    }
}
