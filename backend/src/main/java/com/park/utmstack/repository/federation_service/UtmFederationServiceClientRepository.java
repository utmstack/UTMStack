package com.park.utmstack.repository.federation_service;

import com.park.utmstack.domain.federation_service.UtmFederationServiceClient;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.Optional;


/**
 * Spring Data  repository for the UtmFederationServiceClient entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmFederationServiceClientRepository extends JpaRepository<UtmFederationServiceClient, Long> {

    Optional<UtmFederationServiceClient> findByFsClientToken(String token);
}
