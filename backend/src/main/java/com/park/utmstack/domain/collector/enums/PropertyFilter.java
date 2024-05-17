package com.park.utmstack.domain.collector.enums;

public enum PropertyFilter {
    GROUP("assetGroup.groupName", "UtmCollector");

    private final String propertyName;
    private final String fromTable;

    PropertyFilter(String propertyName, String fromTable) {
        this.propertyName = propertyName;
        this.fromTable = fromTable;
    }

    public String getPropertyName() {
        return propertyName;
    }

    public String getFromTable() {
        return fromTable;
    }
}
