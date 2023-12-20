package com.park.utmstack.domain;


import javax.persistence.*;
import java.io.Serializable;

/**
 * A UtmServerModule.
 */
@Entity
@Table(name = "utm_server_module")
public class UtmServerModule implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(name = "server_id")
    private Long serverId;

    @Column(name = "module_name")
    private String moduleName;

    @Column(name = "needs_restart")
    private Boolean needsRestart;

    @Column(name = "pretty_name")
    private String prettyName;

    @ManyToOne
    @JoinColumn(name = "server_id", referencedColumnName = "id", insertable = false, updatable = false)
    private UtmServer server;

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public Long getServerId() {
        return serverId;
    }

    public UtmServerModule serverId(Long serverId) {
        this.serverId = serverId;
        return this;
    }

    public void setServerId(Long serverId) {
        this.serverId = serverId;
    }

    public String getModuleName() {
        return moduleName;
    }

    public UtmServerModule moduleName(String moduleName) {
        this.moduleName = moduleName;
        return this;
    }

    public void setModuleName(String moduleName) {
        this.moduleName = moduleName;
    }

    public Boolean isNeedsRestart() {
        return needsRestart;
    }

    public UtmServerModule needsRestart(Boolean needsRestart) {
        this.needsRestart = needsRestart;
        return this;
    }

    public void setNeedsRestart(Boolean needsRestart) {
        this.needsRestart = needsRestart;
    }

    public String getPrettyName() {
        return prettyName;
    }

    public UtmServerModule prettyName(String prettyName) {
        this.prettyName = prettyName;
        return this;
    }

    public void setPrettyName(String prettyName) {
        this.prettyName = prettyName;
    }

    public UtmServer getServer() {
        return server;
    }

    public void setServer(UtmServer server) {
        this.server = server;
    }
}
