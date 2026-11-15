package com.park.utmstack.domain.getting_started;

import org.hibernate.annotations.GenericGenerator;

import jakarta.persistence.*;
import org.springframework.data.domain.Persistable;

@Entity
@Table(name = "utm_getting_started")
public class UtmGettingStarted implements Persistable<Long> {
    // Hibernate 6.6 (Boot 3.4.5): track whether loaded from DB so save() uses persist() for
    // new (app-assigned-id) instances instead of merge(), which throws on 6.6.
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
    @Enumerated(EnumType.STRING)
    @Column(name = "step_short", nullable = false, length = 255)
    private GettingStartedStepEnum stepShort;

    @Column(name = "step_order", nullable = false)
    private int stepOrder;

    @Column(name = "completed")
    private boolean completed;

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public GettingStartedStepEnum getStepShort() {
        return stepShort;
    }

    public void setStepShort(GettingStartedStepEnum stepShort) {
        this.stepShort = stepShort;
    }


    public int getStepOrder() {
        return stepOrder;
    }

    public void setStepOrder(int stepOrder) {
        this.stepOrder = stepOrder;
    }

    public boolean isCompleted() {
        return completed;
    }

    public void setCompleted(boolean completed) {
        this.completed = completed;
    }
}
