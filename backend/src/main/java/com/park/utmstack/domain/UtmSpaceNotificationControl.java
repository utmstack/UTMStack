package com.park.utmstack.domain;


import org.hibernate.annotations.GenericGenerator;

import jakarta.persistence.*;
import org.springframework.data.domain.Persistable;
import jakarta.validation.constraints.NotNull;
import java.io.Serializable;
import java.time.Instant;

/**
 * A UtmSpaceNotificationLast.
 */
@Entity
@Table(name = "utm_space_notification_control")
public class UtmSpaceNotificationControl implements Serializable, Persistable<Long>{

    private static final long serialVersionUID = 1L;

    // Hibernate 6.6 (Boot 3.4.5) changed merge() so that save() on an entity whose id is
    // pre-assigned by the app (this app assigns ids from a sequence / fixed value before
    // persisting) throws ObjectOptimisticLockingFailureException. Tracking whether the instance
    // was loaded from the DB (via @PostLoad) makes Spring Data call persist() for new instances
    // and merge() for existing ones - the pre-6.6 behavior.
    private transient boolean isNew = true;

    @jakarta.persistence.PostLoad
    void utmstackOnLoad() {
        this.isNew = false;
    }

    @Override
    public boolean isNew() {
        return this.isNew;
    }


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
