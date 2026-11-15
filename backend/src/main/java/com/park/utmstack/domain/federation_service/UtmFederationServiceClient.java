package com.park.utmstack.domain.federation_service;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.Id;
import jakarta.persistence.Table;

import java.io.Serializable;

/**
 * A UtmFederationServiceClient.
 *
 * <p>Singleton row: the app creates it with {@code id = 1} (see
 * {@code UtmFederationServiceClientService.generateApiToken()}) and reads it back via
 * {@code findById(1L)}.
 *
 * <p>Hibernate 6.6 (Boot 3.4.5): the id is a plain <em>assigned</em> {@code @Id} (no
 * {@code @GeneratedValue}). With a pre-set id, {@code save()} on a fresh instance routes to
 * {@code em.merge()}, which upserts (SELECT id -> absent -> INSERT id=1; present -> UPDATE).
 * The {@code utm_federation_service_client_id_seq} default on the column is harmless: merge
 * writes the assigned id=1 explicitly.
 */
@Entity
@Table(name = "utm_federation_service_client")
public class UtmFederationServiceClient implements Serializable {

    private static final long serialVersionUID = 1L;

    // Assigned identifier (no @GeneratedValue): the app always supplies id = 1 for this row.
    @Id
    private Long id;

    @Column(name = "fs_client_token")
    private String fsClientToken;

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getFsClientToken() {
        return fsClientToken;
    }

    public UtmFederationServiceClient fsClientToken(String fsClientToken) {
        this.fsClientToken = fsClientToken;
        return this;
    }

    public void setFsClientToken(String fsClientToken) {
        this.fsClientToken = fsClientToken;
    }
}
