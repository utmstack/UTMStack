package com.utmstack.userauditor.model.winevent;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Getter;

import java.util.Date;

@Getter
public class EventLog {
    @JsonProperty("@timestamp")
    public String timestamp;
    @JsonProperty("@version")
    public String version;
    public String computer_name;
    public String dataSource;
    public String dataType;
    public Date deviceTime;
    public Global global;
    public String id;
    public Logx logx;
}
