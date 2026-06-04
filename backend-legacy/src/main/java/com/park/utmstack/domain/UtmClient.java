package com.park.utmstack.domain;


import javax.persistence.*;
import java.io.Serializable;
import java.time.Instant;

/**
 * A UtmClient.
 */
@Entity
@Table(name = "utm_client")
public class UtmClient implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(name = "client_name")
    private String clientName;

    @Column(name = "client_domain")
    private String clientDomain;

    @Column(name = "client_prefix")
    private String clientPrefix;

    @Column(name = "client_mail")
    private String clientMail;

    @Column(name = "client_user")
    private String clientUser;

    @Column(name = "client_pass")
    private String clientPass;

    @Column(name = "client_licence_creation")
    private Instant clientLicenceCreation;

    @Column(name = "client_licence_expire")
    private Instant clientLicenceExpire;

    @Column(name = "client_licence_id")
    private String clientLicenceId;

    @Column(name = "client_licence_verified")
    private Boolean clientLicenceVerified;

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getClientName() {
        return clientName;
    }

    public UtmClient clientName(String clientName) {
        this.clientName = clientName;
        return this;
    }

    public void setClientName(String clientName) {
        this.clientName = clientName;
    }

    public String getClientPrefix() {
        return clientPrefix;
    }

    public UtmClient clientPrefix(String clientPrefix) {
        this.clientPrefix = clientPrefix;
        return this;
    }

    public void setClientPrefix(String clientPrefix) {
        this.clientPrefix = clientPrefix;
    }

    public String getClientMail() {
        return clientMail;
    }

    public UtmClient clientMail(String clientMail) {
        this.clientMail = clientMail;
        return this;
    }

    public void setClientMail(String clientMail) {
        this.clientMail = clientMail;
    }

    public String getClientUser() {
        return clientUser;
    }

    public UtmClient clientUser(String clientUser) {
        this.clientUser = clientUser;
        return this;
    }

    public void setClientUser(String clientUser) {
        this.clientUser = clientUser;
    }

    public String getClientPass() {
        return clientPass;
    }

    public UtmClient clientPass(String clientPass) {
        this.clientPass = clientPass;
        return this;
    }

    public void setClientPass(String clientPass) {
        this.clientPass = clientPass;
    }

    public String getClientLicenceId() {
        return clientLicenceId;
    }

    public UtmClient clientLicenceId(String clientLicenceId) {
        this.clientLicenceId = clientLicenceId;
        return this;
    }

    public void setClientLicenceId(String clientLicenceId) {
        this.clientLicenceId = clientLicenceId;
    }

    public Boolean isClientLicenceVerified() {
        return clientLicenceVerified;
    }

    public UtmClient clientLicenceVerified(Boolean clientLicenceVerified) {
        this.clientLicenceVerified = clientLicenceVerified;
        return this;
    }

    public void setClientLicenceVerified(Boolean clientLicenceVerified) {
        this.clientLicenceVerified = clientLicenceVerified;
    }

    public Instant getClientLicenceCreation() {
        return clientLicenceCreation;
    }

    public void setClientLicenceCreation(Instant clientLicenceCreation) {
        this.clientLicenceCreation = clientLicenceCreation;
    }

    public Instant getClientLicenceExpire() {
        return clientLicenceExpire;
    }

    public void setClientLicenceExpire(Instant clientLicenceExpire) {
        this.clientLicenceExpire = clientLicenceExpire;
    }

    public String getClientDomain() {
        return clientDomain;
    }

    public void setClientDomain(String clientDomain) {
        this.clientDomain = clientDomain;
    }
}
