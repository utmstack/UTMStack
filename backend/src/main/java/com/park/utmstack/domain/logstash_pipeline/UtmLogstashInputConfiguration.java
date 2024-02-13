package com.park.utmstack.domain.logstash_pipeline;

import org.hibernate.annotations.GenericGenerator;

import javax.persistence.*;
import javax.validation.constraints.Size;
import java.io.Serializable;

/**
 * A UtmLogstashInputConfiguration.
 */
@Entity
@Table(name = "utm_logstash_input_configuration")
public class UtmLogstashInputConfiguration implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GenericGenerator(name = "CustomIdentityGenerator", strategy = "com.park.utmstack.util.CustomIdentityGenerator")
    @GeneratedValue(generator = "CustomIdentityGenerator")
    @Column(name = "id", updatable = false)
    private Long id;

    @Column(name = "input_id")
    private Integer inputId;

    @Size(max = 50)
    @Column(name = "conf_key")
    private String confKey;

    @Size(max = 255)
    @Column(name = "conf_value")
    private String confValue;

    @Size(max = 50)
    @Column(name = "conf_type")
    private String confType;

    @Column(name = "conf_required")
    private Boolean confRequired;

    @Size(max = 400)
    @Column(name = "conf_validation_regex")
    private String confValidationRegex;

    @Column(name = "system_owner")
    private Boolean systemOwner;

    public UtmLogstashInputConfiguration(){}

    public UtmLogstashInputConfiguration(Long id,
                                         Integer inputId,
                                         String confKey,
                                         String confValue,
                                         String confType,
                                         Boolean confRequired,
                                         String confValidationRegex,
                                         Boolean systemOwner) {
        this.id = id;
        this.inputId = inputId;
        this.confKey = confKey;
        this.confValue = confValue;
        this.confType = confType;
        this.confRequired = confRequired;
        this.confValidationRegex = confValidationRegex;
        this.systemOwner = systemOwner;
    }

    public Long getId() {
        return this.id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public Integer getInputId() {
        return this.inputId;
    }

    public void setInputId(Integer inputId) {
        this.inputId = inputId;
    }

    public String getConfKey() {
        return this.confKey;
    }

    public void setConfKey(String confKey) {
        this.confKey = confKey;
    }

    public String getConfValue() {
        return this.confValue;
    }

    public void setConfValue(String confValue) {
        this.confValue = confValue;
    }

    public String getConfType() {
        return this.confType;
    }

    public void setConfType(String confType) {
        this.confType = confType;
    }

    public Boolean getConfRequired() {
        return this.confRequired;
    }

    public void setConfRequired(Boolean confRequired) {
        this.confRequired = confRequired;
    }

    public String getConfValidationRegex() {
        return this.confValidationRegex;
    }

    public void setConfValidationRegex(String confValidationRegex) {
        this.confValidationRegex = confValidationRegex;
    }

    public Boolean getSystemOwner() {
        return this.systemOwner;
    }

    public void setSystemOwner(Boolean systemOwner) {
        this.systemOwner = systemOwner;
    }
}
