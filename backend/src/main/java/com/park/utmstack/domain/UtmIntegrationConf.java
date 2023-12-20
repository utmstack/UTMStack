package com.park.utmstack.domain;


import javax.persistence.*;
import java.io.Serializable;

/**
 * A UtmIntegrationConf.
 */
@Entity
@Table(name = "utm_integration_conf")
public class UtmIntegrationConf implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(name = "integration_id")
    private Long integrationId;

    @Column(name = "conf_short")
    private String confShort;

    @Column(name = "conf_large")
    private String confLarge;

    @Column(name = "conf_description")
    private String confDescription;

    @Column(name = "conf_value")
    private String confValue;

    @Column(name = "conf_datatype")
    private String confDatatype;

    @ManyToOne
    @JoinColumn(name = "integration_id", referencedColumnName = "id", insertable = false, updatable = false)
    private UtmIntegration integration;

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public Long getIntegrationId() {
        return integrationId;
    }

    public UtmIntegrationConf integrationId(Long integrationId) {
        this.integrationId = integrationId;
        return this;
    }

    public void setIntegrationId(Long integrationId) {
        this.integrationId = integrationId;
    }

    public String getConfShort() {
        return confShort;
    }

    public UtmIntegrationConf confShort(String confShort) {
        this.confShort = confShort;
        return this;
    }

    public void setConfShort(String confShort) {
        this.confShort = confShort;
    }

    public String getConfLarge() {
        return confLarge;
    }

    public UtmIntegrationConf confLarge(String confLarge) {
        this.confLarge = confLarge;
        return this;
    }

    public void setConfLarge(String confLarge) {
        this.confLarge = confLarge;
    }

    public String getConfDescription() {
        return confDescription;
    }

    public UtmIntegrationConf confDescription(String confDescription) {
        this.confDescription = confDescription;
        return this;
    }

    public void setConfDescription(String confDescription) {
        this.confDescription = confDescription;
    }

    public String getConfValue() {
        return confValue;
    }

    public UtmIntegrationConf confValue(String confValue) {
        this.confValue = confValue;
        return this;
    }

    public void setConfValue(String confValue) {
        this.confValue = confValue;
    }

    public String getConfDatatype() {
        return confDatatype;
    }

    public UtmIntegrationConf confDatatype(String confDatatype) {
        this.confDatatype = confDatatype;
        return this;
    }

    public void setConfDatatype(String confDatatype) {
        this.confDatatype = confDatatype;
    }

    public UtmIntegration getIntegration() {
        return integration;
    }

    public void setIntegration(UtmIntegration integration) {
        this.integration = integration;
    }
}
