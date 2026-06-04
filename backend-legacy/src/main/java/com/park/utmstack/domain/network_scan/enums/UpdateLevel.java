package com.park.utmstack.domain.network_scan.enums;

/**
 * Identify who was the one who update an asset. There is a level of priority between them:
 * AGENT overwrite SCANNER and SCANNER overwrite DATASOURCE
 */
public enum UpdateLevel {
    DATASOURCE,
    SCANNER,
    AGENT
}
