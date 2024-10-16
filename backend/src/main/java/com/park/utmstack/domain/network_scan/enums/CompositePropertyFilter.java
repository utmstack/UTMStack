package com.park.utmstack.domain.network_scan.enums;

import com.park.utmstack.domain.network_scan.Property;
import lombok.Getter;
import lombok.RequiredArgsConstructor;

@RequiredArgsConstructor
public enum CompositePropertyFilter implements Property {

    RULE_DATA_TYPES(new String[]{"u0.dataType"}, "UtmCorrelationRules", new String[]{"p.dataTypes"}),
    DATA_TYPES(new String[]{"u0.dataType", "u1.dataType"}, "UtmNetworkScan", new String[]{"p.dataInputSourceList", "p.dataInputIpList"});

    private final String[] propertyNames;

    @Getter
    private final String fromTable;

    private final String[] joinTables;

    @Override
    public String getPropertyName() {
        if(propertyNames.length > 1){
            return "COALESCE(" + String.join(", ", propertyNames) + ")";
        } else {
            return propertyNames[0];
        }

    }

    @Override
    public String getJoinTable() {
        StringBuilder joinClause = new StringBuilder();

        for (int i = 0; i < joinTables.length; i++) {
            joinClause.append(" LEFT JOIN ")
                    .append(joinTables[i])
                    .append(" u")
                    .append(i);
        }

        return joinClause.toString();
    }
}
