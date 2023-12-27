package com.park.utmstack.domain.network_scan.enums;

public enum PropertyFilter {
    IP("assetIp", "UtmNetworkScan"),
    MAC("assetMac", "UtmNetworkScan"),
    ALIAS("assetAlias", "UtmNetworkScan"),
    NAME("assetName", "UtmNetworkScan"),
    OS("assetOs", "UtmNetworkScan"),
    ALIVE("assetAlive", "UtmNetworkScan"),
    STATUS("assetStatus", "UtmNetworkScan"),
    PROBE("serverName", "UtmNetworkScan"),
    TYPE("assetType.typeName", "UtmNetworkScan"),
    SEVERITY("assetSeverity", "UtmNetworkScan"),
    PORTS("port", "UtmPorts"),
    GROUP("assetGroup.groupName", "UtmNetworkScan");

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
