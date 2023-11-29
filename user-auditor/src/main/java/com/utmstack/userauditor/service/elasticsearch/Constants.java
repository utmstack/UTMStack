package com.utmstack.userauditor.service.elasticsearch;

public final class Constants {
    
    // ----------------------------------------------------------------------------------
    // - Index date used format
    // ----------------------------------------------------------------------------------
    public static final String INDEX_TIMESTAMP_FORMAT = "strict_date_optional_time_nanos";

    // ----------------------------------------------------------------------------------
    // - Indices common fields
    // ----------------------------------------------------------------------------------

    public static final String logxWineventlogEventDataSubjectUserSidKeyword = "logx.wineventlog.event_data.SubjectUserSid.keyword";
    public static final String logxWineventlogEventDataTargetUserSidKeyword = "logx.wineventlog.event_data.TargetUserSid.keyword";
    public static final String logxWineventlogEventDataMemberSidKeyword = "logx.wineventlog.event_data.MemberSid.keyword";

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
