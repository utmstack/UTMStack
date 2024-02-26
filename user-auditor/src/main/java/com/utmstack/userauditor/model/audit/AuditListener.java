package com.utmstack.userauditor.model.audit;

import com.utmstack.userauditor.model.Audit;

import javax.persistence.*;
import java.time.LocalDateTime;

public class AuditListener {

    @PrePersist
    public void setCreatedOn(Auditable auditable) {
        Audit audit = auditable.getAudit();

        if(audit == null) {
            audit = new Audit();
            auditable.setAudit(audit);
        }

        audit.setCreatedDate(LocalDateTime.now());
    }

    @PreUpdate
    public void setUpdatedOn(Auditable auditable) {
        Audit audit = auditable.getAudit();

        audit.setModifiedDate(LocalDateTime.now());
    }
}
