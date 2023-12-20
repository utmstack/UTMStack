package com.park.utmstack.domain.agents_manager;


import javax.persistence.*;
import javax.validation.constraints.NotNull;
import javax.validation.constraints.Size;
import java.io.Serializable;

/**
 * A UtmAgentManager.
 */
@Entity
@Table(name = "utm_agent_manager")
public class UtmAgentManager implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @NotNull
    @Size(max = 100)
    @Column(name = "manager_host", length = 100, nullable = false, unique = true)
    private String managerHost;

    @NotNull
    @Size(max = 50)
    @Column(name = "manager_version", length = 50, nullable = false)
    private String managerVersion;

    @NotNull
    @Column(name = "manager_port", nullable = false)
    private Integer managerPort;

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getManagerHost() {
        return managerHost;
    }

    public UtmAgentManager managerHost(String managerHost) {
        this.managerHost = managerHost;
        return this;
    }

    public void setManagerHost(String managerHost) {
        this.managerHost = managerHost;
    }

    public String getManagerVersion() {
        return managerVersion;
    }

    public UtmAgentManager managerVersion(String managerVersion) {
        this.managerVersion = managerVersion;
        return this;
    }

    public void setManagerVersion(String managerVersion) {
        this.managerVersion = managerVersion;
    }

    public Integer getManagerPort() {
        return managerPort;
    }

    public UtmAgentManager managerPort(Integer managerPort) {
        this.managerPort = managerPort;
        return this;
    }

    public void setManagerPort(Integer managerPort) {
        this.managerPort = managerPort;
    }

    @Override
    public String toString() {
        return String.format("https://%1$s:%2$s/api/%3$s/agent", managerHost, managerPort, managerVersion);
    }
}
