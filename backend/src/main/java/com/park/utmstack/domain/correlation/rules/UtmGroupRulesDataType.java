package com.park.utmstack.domain.correlation.rules;

import javax.persistence.*;
import java.io.Serializable;
import java.time.Clock;
import java.time.Instant;

/**
 * UtmGroupRulesDataType entity template.
 */
@Entity
@Table(name = "utm_group_rules_data_type")
public class UtmGroupRulesDataType implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    @Column(name = "id", nullable = false)
    private Long id;

    @Column(name = "rule_id", nullable = false)
    private Long ruleId;

    @Column(name = "data_type_id", nullable = false)
    private Long dataTypeId;

    @Column(name = "last_update", nullable = false)
    private Instant lastUpdate;

    public UtmGroupRulesDataType(Long id, Long ruleId, Long dataTypeId, Instant lastUpdate) {
        this.id = id;
        this.ruleId = ruleId;
        this.dataTypeId = dataTypeId;
        this.lastUpdate = lastUpdate;
    }

    public UtmGroupRulesDataType() {
    }

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public Long getRuleId() {
        return ruleId;
    }

    public void setRuleId(Long ruleId) {
        this.ruleId = ruleId;
    }

    public Long getDataTypeId() {
        return dataTypeId;
    }

    public void setDataTypeId(Long dataTypeId) {
        this.dataTypeId = dataTypeId;
    }

    public Instant getLastUpdate() {
        return lastUpdate;
    }

    public void setLastUpdate() {
        this.lastUpdate = Instant.now(Clock.systemUTC());
    }
}
