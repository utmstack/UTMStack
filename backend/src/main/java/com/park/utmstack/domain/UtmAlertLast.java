package com.park.utmstack.domain;


import jakarta.persistence.*;
import java.io.Serializable;
import java.time.Instant;
import java.time.ZoneOffset;
import java.util.Objects;

/**
 * A UtmAlertLast.
 *
 * <p>This is a singleton row that the app always creates with {@code id = 1} and reads back
 * via {@code findById(1L)}; the {@code utm_alert_last.id} column has no DB default.
 *
 * <p>Hibernate 6.6 (Boot 3.4.5): the id is a plain <em>assigned</em> {@code @Id} (no
 * {@code @GeneratedValue}). The pre-migration generated-id mapping rejected a pre-set
 * {@code id = 1} under {@code persist()} ("detached entity passed to persist"). Without a
 * generator, {@code save()} on a fresh instance (non-null id) routes to {@code em.merge()},
 * which upserts (SELECT id -> absent -> INSERT id=1; present -> UPDATE) - the correct
 * behavior for this singleton.
 */
@Entity
@Table(name = "utm_alert_last")
public class UtmAlertLast implements Serializable {

    private static final long serialVersionUID = 1L;

    // Assigned identifier (no @GeneratedValue): the app always supplies id = 1 for this row.
    @Id
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
