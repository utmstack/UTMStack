package com.park.utmstack.domain;


import org.hibernate.annotations.GenericGenerator;

import javax.persistence.*;
import javax.validation.constraints.NotNull;
import java.io.Serializable;
import java.time.Instant;

/**
 * A UtmSpaceNotificationLast.
 */
@Entity
@Table(name = "utm_space_notification_control")
public class UtmSpaceNotificationControl implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GenericGenerator(name = "CustomIdentityGenerator", strategy = "com.park.utmstack.util.CustomIdentityGenerator")
    @GeneratedValue(generator = "CustomIdentityGenerator")
    private Long id;

    @NotNull
    @Column(name = "next_notification", nullable = false)
    private Instant nextNotification;

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public Instant getNextNotification() {
        return nextNotification;
    }

    public void setNextNotification(Instant nextNotification) {
        this.nextNotification = nextNotification;
    }
}
