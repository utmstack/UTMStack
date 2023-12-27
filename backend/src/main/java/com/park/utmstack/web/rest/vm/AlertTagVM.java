package com.park.utmstack.web.rest.vm;

import com.park.utmstack.domain.UtmAlertTag;

public class AlertTagVM {
    private Long id;
    private String tagName;
    private String tagColor;
    private Boolean systemOwner;

    public AlertTagVM() {
    }

    public AlertTagVM(UtmAlertTag tag) {
        this.id = tag.getId();
        this.tagName = tag.getTagName();
        this.tagColor = tag.getTagColor();
        this.systemOwner = tag.getSystemOwner();
    }

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getTagName() {
        return tagName;
    }

    public void setTagName(String tagName) {
        this.tagName = tagName;
    }

    public String getTagColor() {
        return tagColor;
    }

    public void setTagColor(String tagColor) {
        this.tagColor = tagColor;
    }

    public Boolean getSystemOwner() {
        return systemOwner;
    }

    public void setSystemOwner(Boolean systemOwner) {
        this.systemOwner = systemOwner;
    }
}
