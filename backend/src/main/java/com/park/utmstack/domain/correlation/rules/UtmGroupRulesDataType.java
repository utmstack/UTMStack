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
    @Column(name = "rule_id", nullable = false)
    private Long ruleId;

    @Id
    @Column(name = "data_type_id", nullable = false)
    private Long dataTypeId;

    @Column(name = "last_update", nullable = false)
    private Instant lastUpdate;

    public UtmGroupRulesDataType(Long id, Long ruleId, Long dataTypeId, Instant lastUpdate) {
        this.ruleId = ruleId;
        this.dataTypeId = dataTypeId;
        this.lastUpdate = lastUpdate;
    }

    public UtmGroupRulesDataType() {
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

    public void setLastUpdate(Instant lastUpdate) {
        this.lastUpdate = lastUpdate;
    }
}
