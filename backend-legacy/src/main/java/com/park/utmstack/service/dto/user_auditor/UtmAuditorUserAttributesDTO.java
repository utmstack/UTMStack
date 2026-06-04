package com.park.utmstack.service.dto.user_auditor;

import java.time.Instant;

/**
 * A UtmAuditorUserAttributes.
 */

public class UtmAuditorUserAttributesDTO {

    private Long id;

    private String attributeKey;

    private String attributeValue;


    private Instant createdDate;

    private Instant modifiedDate;

    UtmAuditorUsersDTO utmAuditorUser;

    public Long getId() {
        return this.id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getAttributeKey() {
        return this.attributeKey;
    }

    public void setAttributeKey(String attributeKey) {
        this.attributeKey = attributeKey;
    }

    public String getAttributeValue() {
        return this.attributeValue;
    }

    public void setAttributeValue(String attributeValue) {
        this.attributeValue = attributeValue;
    }

    public Instant getCreatedDate() {
        return this.createdDate;
    }

    public void setCreatedDate(Instant createdDate) {
        this.createdDate = createdDate;
    }

    public Instant getModifiedDate() {
        return this.modifiedDate;
    }

    public void setModifiedDate(Instant modifiedDate) {
        this.modifiedDate = modifiedDate;
    }

    public UtmAuditorUsersDTO getUtmAuditorUser() { return utmAuditorUser; }

    public void setUtmAuditorUser(UtmAuditorUsersDTO utmAuditorUser) { this.utmAuditorUser = utmAuditorUser; }
}
