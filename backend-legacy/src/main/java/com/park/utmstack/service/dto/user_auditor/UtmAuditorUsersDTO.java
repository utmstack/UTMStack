package com.park.utmstack.service.dto.user_auditor;

import java.util.List;

/**
 * A UtmAuditorUsers.
 */

public class UtmAuditorUsersDTO {

    private String name;
    private String sid;
    private UtmAuditorUserSourcesDTO source;
    private List<UtmAuditorUserAttributesDTO> attributes;

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getSid() {
        return sid;
    }

    public void setSid(String sid) {
        this.sid = sid;
    }

    public UtmAuditorUserSourcesDTO getSource() {
        return source;
    }

    public void setSource(UtmAuditorUserSourcesDTO source) {
        this.source = source;
    }

    public List<UtmAuditorUserAttributesDTO> getAttributes() {
        return attributes;
    }

    public void setAttributes(List<UtmAuditorUserAttributesDTO> attributes) {
        this.attributes = attributes;
    }
}
