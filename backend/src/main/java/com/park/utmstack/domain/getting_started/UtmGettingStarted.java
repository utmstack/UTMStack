package com.park.utmstack.domain.getting_started;

import org.hibernate.annotations.GenericGenerator;

import javax.persistence.*;

@Entity
@Table(name = "utm_getting_started")
public class UtmGettingStarted {
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
