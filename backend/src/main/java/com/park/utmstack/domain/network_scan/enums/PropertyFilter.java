package com.park.utmstack.domain.network_scan.enums;

import com.park.utmstack.domain.network_scan.Property;
import lombok.Getter;
import lombok.RequiredArgsConstructor;

@Getter
@RequiredArgsConstructor
public enum PropertyFilter implements Property {
    IP("assetIp", "UtmNetworkScan", ""),
    MAC("assetMac", "UtmNetworkScan", ""),
    ALIAS("assetAlias", "UtmNetworkScan", ""),
    NAME("assetName", "UtmNetworkScan", ""),
    OS("assetOs", "UtmNetworkScan", ""),
    ALIVE("assetAlive", "UtmNetworkScan", ""),
    STATUS("assetStatus", "UtmNetworkScan", ""),
    PROBE("serverName", "UtmNetworkScan", ""),
    TYPE("assetType.typeName", "UtmNetworkScan", ""),
    SEVERITY("assetSeverity", "UtmNetworkScan", ""),
    PORTS("port", "UtmPorts", ""),
    GROUP("assetGroup.groupName", "UtmNetworkScan", ""),
    COLLECTOR_IP("ip", "UtmCollector", ""),
    COLLECTOR_GROUP("assetGroup.groupName", "UtmCollector", ""),
    RULE_TECHNIQUE("ruleTechnique","UtmCorrelationRules", ""),
    RULE_CATEGORY("ruleCategory","UtmCorrelationRules", "");
    //RULE_DATA_TYPES("jt.dataType","UtmCorrelationRules", "dataTypes");
    //DATA_TYPES("jt.dataType","UtmNetworkScan", "dataInputSourceList");

    private final String propertyName;
    private final String fromTable;
    private final String joinTable;
}
