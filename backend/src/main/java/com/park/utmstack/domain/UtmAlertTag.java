package com.park.utmstack.domain;


import javax.persistence.*;
import javax.validation.constraints.NotBlank;
import javax.validation.constraints.Size;
import java.io.Serializable;

@Entity
@Table(name = "utm_alert_tag")
public class UtmAlertTag implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @NotBlank
    @Size(max = 50)
    @Column(name = "tag_name", length = 50, nullable = false, unique = true)
    private String tagName;

    @Size(max = 15)
    @Column(name = "tag_color", length = 15)
    private String tagColor;

    @Column(name = "system_owner", length = 15)
    private Boolean systemOwner;

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
