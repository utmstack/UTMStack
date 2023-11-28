package com.utmstack.userauditor.service.interfaces;

import com.utmstack.userauditor.model.winevent.EventLog;
import com.utmstack.userauditor.model.UserSource;
import com.utmstack.userauditor.service.type.SourceType;

import java.util.List;
import java.util.Map;

public interface Source {
    Map<String, List<EventLog>> findUsers(UserSource userSource) throws Exception;
    SourceType getType();
}
