package com.utmstack.userauditor.model.audit;

import com.utmstack.userauditor.model.Audit;

public interface Auditable {
    Audit getAudit();
    void setAudit(Audit audit);
}
