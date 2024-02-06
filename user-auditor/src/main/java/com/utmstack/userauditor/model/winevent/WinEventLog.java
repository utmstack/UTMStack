package com.utmstack.userauditor.model.winevent;

import com.fasterxml.jackson.annotation.JsonProperty;

import java.util.ArrayList;

public class WinEventLog{
    public String activity_id;
    public Agent agent;
    public Beat beat;
    public Ecs ecs;
    public Event event;
    @JsonProperty("event_data")
    public EventData eventData;
    @JsonProperty("event_id")
    public int eventId;
    public String event_name;
    public Host host;
    public ArrayList<String> keywords;
    public String level;
    public String log_name;

    public String message;
    public String opcode;
    public String process_id;
    public String provider_guid;
    public String record_number;
    public String source_name;
    public ArrayList<String> tags;
    public String task;
    public String thread_id;
    public String version;
}


