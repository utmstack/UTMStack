package com.park.utmstack.service.dto.auditable;

import java.util.Map;

public interface AuditableDTO {
    Map<String, Object> toAuditMap();
}
