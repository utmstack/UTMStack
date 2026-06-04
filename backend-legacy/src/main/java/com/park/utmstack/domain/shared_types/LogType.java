package com.park.utmstack.domain.shared_types;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;
import com.park.utmstack.util.MapUtil;
import org.springframework.util.StringUtils;

import java.time.Instant;
import java.time.ZoneId;
import java.time.format.DateTimeFormatter;
import java.util.Locale;
import java.util.Map;

@JsonInclude(JsonInclude.Include.NON_NULL)
@JsonIgnoreProperties(ignoreUnknown = true)
public class LogType {
    private String id;
    @JsonProperty("@timestamp")
    private String timestamp;
    private String dataSource;
    private String dataType;
    private Map<String, Object> logx;

    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }

    public String getTimestamp() {
        return timestamp;
    }

    public String getTimestampFormatted() {
        try {
            return StringUtils.hasText(timestamp) ? DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss").withLocale(Locale.getDefault()).withZone(
                ZoneId.systemDefault()).format(Instant.parse(timestamp)) : null;
        } catch (Exception e) {
            return null;
        }
    }

    public void setTimestamp(String timestamp) {
        this.timestamp = timestamp;
    }

    public String getDataSource() {
        return dataSource;
    }

    public void setDataSource(String dataSource) {
        this.dataSource = dataSource;
    }

    public String getDataType() {
        return dataType;
    }

    public void setDataType(String dataType) {
        this.dataType = dataType;
    }

    public Map<String, Object> getLogx() {
        return logx;
    }

    public void setLogx(Map<String, Object> logx) {
        this.logx = logx;
    }

    public Map<String, String> getLogxFlatted() {
        return MapUtil.flattenToStringMap(logx, true);
    }
}
