package com.park.utmstack.domain.correlation.rules;

import com.park.utmstack.domain.correlation.config.UtmDataTypes;

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
    @EmbeddedId
    private UtmGroupRulesDataTypeKey id;

    @Column(name = "last_update", nullable = false)
    private Instant lastUpdate;

    @ManyToOne(fetch = FetchType.LAZY)
    @MapsId("ruleId")
    @JoinColumn(name = "rule_id", insertable = false, updatable = false)
    private UtmCorrelationRules rule;

    @ManyToOne(fetch = FetchType.LAZY)
    @MapsId("dataTypeId")
    @JoinColumn(name = "data_type_id", insertable = false, updatable = false)
    private UtmDataTypes dataType;
}
