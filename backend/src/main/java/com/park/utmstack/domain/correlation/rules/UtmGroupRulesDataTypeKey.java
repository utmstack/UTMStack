package com.park.utmstack.domain.correlation.rules;

import javax.persistence.Embeddable;
import java.io.Serializable;
import java.util.Objects;

@Embeddable
public class UtmGroupRulesDataTypeKey implements Serializable {
    private Long ruleId;
    private Long dataTypeId;

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (!(o instanceof UtmGroupRulesDataTypeKey that)) return false;
        return Objects.equals(ruleId, that.ruleId) &&
                Objects.equals(dataTypeId, that.dataTypeId);
    }

    @Override
    public int hashCode() {
        return Objects.hash(ruleId, dataTypeId);
    }
}

