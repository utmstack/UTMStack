package com.park.utmstack.domain.logstash_pipeline;

import org.hibernate.annotations.GenericGenerator;

import javax.persistence.*;
import javax.validation.constraints.Size;
import java.io.Serializable;

/**
 * A UtmLogstashInput.
 */
@Entity
@Table(name = "utm_logstash_input")
public class UtmLogstashInput implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GenericGenerator(name = "CustomIdentityGenerator", strategy = "com.park.utmstack.util.CustomIdentityGenerator")
    @GeneratedValue(generator = "CustomIdentityGenerator")
    @Column(name = "id", updatable = false)
    private Long id;

    @Column(name = "pipeline_id")
    private Integer pipelineId;

    @Size(max = 250)
    @Column(name = "input_pretty_name", length = 250)
    private String inputPrettyName;

    @Size(max = 100)
    @Column(name = "input_plugin", length = 100)
    private String inputPlugin;

    @Column(name = "input_with_ssl")
    private Boolean inputWithSsl;

    @Column(name = "system_owner")
    private Boolean systemOwner;

    public UtmLogstashInput(){}

    public UtmLogstashInput(Long id,
                            Integer pipelineId,
                            String inputPrettyName,
                            String inputPlugin,
                            Boolean inputWithSsl, Boolean systemOwner) {
        this.id = id;
        this.pipelineId = pipelineId;
        this.inputPrettyName = inputPrettyName;
        this.inputPlugin = inputPlugin;
        this.inputWithSsl = inputWithSsl;
        this.systemOwner = systemOwner;
    }

    public Long getId() {
        return this.id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public Integer getPipelineId() {
        return pipelineId;
    }

    public void setPipelineId(Integer pipelineId) {
        this.pipelineId = pipelineId;
    }

    public String getInputPrettyName() {
        return this.inputPrettyName;
    }

    public void setInputPrettyName(String inputPrettyName) {
        this.inputPrettyName = inputPrettyName;
    }

    public String getInputPlugin() {
        return this.inputPlugin;
    }

    public void setInputPlugin(String inputPlugin) {
        this.inputPlugin = inputPlugin;
    }

    public Boolean getInputWithSsl() {
        return this.inputWithSsl;
    }

    public void setInputWithSsl(Boolean inputWithSsl) {
        this.inputWithSsl = inputWithSsl;
    }

    public Boolean getSystemOwner() {
        return this.systemOwner;
    }

    public void setSystemOwner(Boolean systemOwner) {
        this.systemOwner = systemOwner;
    }
}
