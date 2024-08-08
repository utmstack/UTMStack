package com.park.utmstack.config;

import com.park.utmstack.domain.index_pattern.enums.SystemIndexPattern;

import java.util.HashMap;
import java.util.Map;

public final class Constants {

    // Regex for acceptable logins
    public static final String LOGIN_REGEX = "^[_.@A-Za-z0-9-]*$";
    public static final String SYSTEM_ACCOUNT = "system";
    public static final String ANONYMOUS_USER = "anonymoususer";
    public static final String ADMIN_USER = "admin";
    public static final String FS_USER = "fsclient";
    public static final String DEFAULT_LANGUAGE = "en";
    public static final String PROP_MAIL_HOST = "utmstack.mail.host";
    public static final String PROP_MAIL_PORT = "utmstack.mail.port";
    public static final String PROP_MAIL_PASSWORD = "utmstack.mail.password";
    public static final String PROP_MAIL_USERNAME = "utmstack.mail.username";
    public static final String PROP_MAIL_SMTP_AUTH = "utmstack.mail.properties.mail.smtp.auth";

    public static final String PROP_EMAIL_PROTOCOL_VALUE = "smtp";
    public static final Integer PROP_EMAIL_PORT_TLS_VALUE = 587;
    public static final Integer PROP_EMAIL_PORT_SSL_VALUE = 465;
    public static final Integer PROP_EMAIL_PORT_NONE_VALUE = 25;
    public static final String PROP_MAIL_FROM = "utmstack.mail.from";
    public static final String PROP_MAIL_BASE_URL = "utmstack.mail.baseUrl";
    public static final String PROP_TWILIO_ACCOUNT_SID = "utmstack.twilio.accountSid";
    public static final String PROP_TWILIO_AUTH_TOKEN = "utmstack.twilio.authToken";
    public static final String PROP_TWILIO_NUMBER = "utmstack.twilio.number";
    public static final String PROP_ALERT_ADDRESS_TO_NOTIFY_ALERTS = "utmstack.alert.addressToNotifyAlerts";
    public static final String PROP_ALERT_ADDRESS_TO_NOTIFY_INCIDENTS = "utmstack.alert.addressToNotifyIncidents";
    public static final String PROP_DATE_SETTINGS_TIMEZONE = "utmstack.time.zone";
    public static final String PROP_NETWORK_SCAN_API_URL = "utmstack.networkScan.apiUrl";
    public static final String PROP_TFA_ENABLE = "utmstack.tfa.enable";

    // ----------------------------------------------------------------------------------
    // - Modules checks
    // ----------------------------------------------------------------------------------
    public static final String MODULE_CHECK_WINEVENTLOG = "Windows events verification";

    // ----------------------------------------------------------------------------------
    // - Application constant
    // ----------------------------------------------------------------------------------
    public static final Map<SystemIndexPattern, String> SYS_INDEX_PATTERN = new HashMap<>();
    public static final Map<String, String> CFG = new HashMap<>();

    // ----------------------------------------------------------------------------------
    // - Index date used format
    // ----------------------------------------------------------------------------------
    public static final String INDEX_TIMESTAMP_FORMAT = "strict_date_optional_time_nanos";

    // ----------------------------------------------------------------------------------
    // - Indices common fields
    // ----------------------------------------------------------------------------------
    public static final String timestamp = "@timestamp";
    public static final String _id = "_id";
    public static final String dataTypeKeyword = "dataType.keyword";


    // ----------------------------------------------------------------------------------
    // - Logs index common fields
    // ----------------------------------------------------------------------------------
    public static final String logxWineventlogLogNameKeyword = "logx.wineventlog.log_name.keyword";
    public static final String logxWineventlogEventNameKeyword = "logx.wineventlog.event_name.keyword";

    // ----------------------------------------------------------------------------------
    // - Alert index common fields
    // ----------------------------------------------------------------------------------
    public static final String alertIdKeyword = "id.keyword";
    public static final String alertStatus = "status";
    public static final String alertIsIncident = "isIncident";
    public static final String alertNameKeyword = "name.keyword";
    public static final String alertSeverityLabel = "severityLabel.keyword";
    public static final String alertCategoryKeyword = "category.keyword";
    public static final String alertDataSourceKeyword = "dataSource.keyword";
    public static final int LOG_ANALYZER_TOTAL_RESULTS = 10000;

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
    public static final String ENV_ENCRYPTION_KEY = "ENCRYPTION_KEY";
    public static final String ENV_GRPC_AGENT_MANAGER_HOST = "GRPC_AGENT_MANAGER_HOST";
    public static final String ENV_GRPC_AGENT_MANAGER_PORT = "GRPC_AGENT_MANAGER_PORT";
    public static final String ENV_INTERNAL_KEY = "INTERNAL_KEY";
    public static final String ENV_SERVER_NAME = "SERVER_NAME";
    public static final String ENV_AD_AUDIT_SERVICE = "AD_AUDIT_SERVICE";

    /**
     * Index policies states
     */
    public static final String STATE_INGEST = "ingest";
    public static final String STATE_OPEN = "open";
    public static final String STATE_DELETE = "delete";
    public static final String STATE_BACKUP = "backup";
    public static final String STATE_SAFE_DELETE = "safe_delete";

    public static final Long DATE_FORMAT_SETTING_ID = 5L;

    // GRPC
    public static final String AGENT_MANAGER_INTERNAL_KEY_HEADER = "internal-key";
    public static final String AGENT_HEADER = "Agent";

    /**
     * Alert Response Rules Execution
     */
    public static final int IRA_EXECUTION_RETRIES = 5;

    // ----------------------------------------------------------------------------------
    // - Constants used to PDF microservice generation
    // ----------------------------------------------------------------------------------
    public static final String FRONT_BASE_URL = "https://10.21.199.3";
    public static final String PDF_SERVICE_URL = "http://web-pdf:8080/generate-pdf";

    private Constants() {
    }
}
