package com.park.utmstack.service.dto.user_auditor;

/**
 * A UtmAuditorUserSourcesDTO.
 */

public class UtmAuditorUserSourcesDTO {

    private Long id;

    private String indexPattern;

    private String indexName;

    private Boolean isActive;


    public Long getId() {
        return this.id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getIndexPattern() {
        return this.indexPattern;
    }

    public void setIndexPattern(String indexPattern) {
        this.indexPattern = indexPattern;
    }

    public String getIndexName() {
        return this.indexName;
    }

    public void setIndexName(String indexName) {
        this.indexName = indexName;
    }

    public Boolean getIsActive() {
        return this.isActive;
    }

    public void setActive(Boolean active) {
        this.isActive = active;
    }
}
