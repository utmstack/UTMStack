package com.park.utmstack.domain.logstash_pipeline;

import javax.persistence.*;
import javax.validation.constraints.Size;
import java.io.Serializable;

/**
 * A UtmPortsConfiguration.
 */
@Entity
@Table(name = "utm_logstash_ports_configuration")
public class UtmPortsConfiguration implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    @Column(name = "id")
    private Long id;

    @Size(max = 50)
    @Column(name = "protocol", length = 50)
    private String protocol;

    @Size(max = 50)
    @Column(name = "port_or_range", length = 50)
    private String portOrRange;

    @Column(name = "system_owner")
    private Boolean systemOwner;

    public UtmPortsConfiguration(){}
    public UtmPortsConfiguration(Long id, String protocol, String portOrRange, Boolean systemOwner) {
        this.id = id;
        this.protocol = protocol;
        this.portOrRange = portOrRange;
        this.systemOwner = systemOwner;
    }

    public Long getId() {
        return this.id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getProtocol() {
        return this.protocol;
    }

    public void setProtocol(String protocol) {
        this.protocol = protocol;
    }

    public String getPortOrRange() {
        return this.portOrRange;
    }

    public void setPortOrRange(String portOrRange) {
        this.portOrRange = portOrRange;
    }

    public Boolean getSystemOwner() {
        return this.systemOwner;
    }

    public void setSystemOwner(Boolean systemOwner) {
        this.systemOwner = systemOwner;
    }
}
