package com.park.utmstack.domain;


import javax.persistence.*;
import java.io.Serializable;

/**
 * A UtmIntegration.
 */
@Entity
@Table(name = "utm_integration")
public class UtmIntegration implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(name = "module_id")
    private Long moduleId;

    @Column(name = "integration_name")
    private String integrationName;

    @Column(name = "integration_description")
    private String integrationDescription;

    @Column(name = "url")
    private String url;

    @Column(name = "integration_icon_path")
    private String integrationIconPath;

    @ManyToOne
    @JoinColumn(name = "module_id", referencedColumnName = "id", insertable = false, updatable = false)
    private UtmServerModule module;

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public Long getModuleId() {
        return moduleId;
    }

    public UtmIntegration moduleId(Long moduleId) {
        this.moduleId = moduleId;
        return this;
    }

    public void setModuleId(Long moduleId) {
        this.moduleId = moduleId;
    }

    public String getIntegrationName() {
        return integrationName;
    }

    public UtmIntegration integrationName(String integrationName) {
        this.integrationName = integrationName;
        return this;
    }

    public void setIntegrationName(String integrationName) {
        this.integrationName = integrationName;
    }

    public String getIntegrationDescription() {
        return integrationDescription;
    }

    public UtmIntegration integrationDescription(String integrationDescription) {
        this.integrationDescription = integrationDescription;
        return this;
    }

    public void setIntegrationDescription(String integrationDescription) {
        this.integrationDescription = integrationDescription;
    }

    public String getUrl() {
        return url;
    }

    public UtmIntegration url(String url) {
        this.url = url;
        return this;
    }

    public void setUrl(String url) {
        this.url = url;
    }

    public UtmServerModule getModule() {
        return module;
    }

    public void setModule(UtmServerModule module) {
        this.module = module;
    }

    public String getIntegrationIconPath() {
        return integrationIconPath;
    }

    public void setIntegrationIconPath(String integrationIconPath) {
        this.integrationIconPath = integrationIconPath;
    }
}
