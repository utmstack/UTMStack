package com.park.utmstack.util.chart_builder;

import com.fasterxml.jackson.annotation.JsonInclude;
import org.opensearch.client.opensearch.cat.indices.IndicesRecord;
import org.springframework.util.StringUtils;

import java.io.Serializable;

@JsonInclude(JsonInclude.Include.NON_NULL)
public class IndexType implements Serializable {
    private String health;
    private String status;
    private String index;
    private Long docsCount;
    private String size;
    private String creationDate;

    public IndexType() {
    }

    public IndexType(IndicesRecord record) {
        this.health = record.health();
        this.status = record.status();
        this.index = record.index();
        this.docsCount = StringUtils.hasText(record.docsCount()) ? Long.parseLong(record.docsCount()) : 0;
        this.size = record.storeSize();
        this.creationDate = record.creationDateString();
    }

    public String getHealth() {
        return health;
    }

    public void setHealth(String health) {
        this.health = health;
    }

    public String getStatus() {
        return status;
    }

    public void setStatus(String status) {
        this.status = status;
    }

    public String getIndex() {
        return index;
    }

    public void setIndex(String index) {
        this.index = index;
    }

    public Long getDocsCount() {
        return docsCount;
    }

    public void setDocsCount(Long docsCount) {
        this.docsCount = docsCount;
    }

    public String getSize() {
        return size;
    }

    public void setSize(String size) {
        this.size = size;
    }

    public String getCreationDate() {
        return creationDate;
    }

    public void setCreationDate(String creationDate) {
        this.creationDate = creationDate;
    }
}
