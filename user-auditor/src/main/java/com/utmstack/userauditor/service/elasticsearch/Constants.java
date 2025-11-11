package com.utmstack.userauditor.service.elasticsearch;

public final class Constants {
    
    // ----------------------------------------------------------------------------------
    // - Index date used format
    // ----------------------------------------------------------------------------------
    public static final String INDEX_TIMESTAMP_FORMAT = "strict_date_optional_time_nanos";

    // ----------------------------------------------------------------------------------
    // - Indices common fields
    // ----------------------------------------------------------------------------------

    public static final String LOG_WINLOG_EVENT_DATA_TARGET_USER_SID_KEYWORD = "log.winlogEventDataTargetUserSid.keyword";

    /**
     * Environment variables
     */
    public static final String ENV_ELASTICSEARCH_HOST = "ELASTICSEARCH_HOST";
    public static final String ENV_ELASTICSEARCH_PORT = "ELASTICSEARCH_PORT";
    public static final String ENV_DB_HOST = "DB_HOST";
    public static final String ENV_DB_PORT = "DB_PORT";
    public static final String ENV_DB_NAME = "DB_NAME";
    public static final String ENV_DB_PASS = "DB_PASS";
    public static final String ENV_DB_USER = "DB_USER";


    private Constants() {
    }
}
