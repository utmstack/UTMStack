package com.park.utmstack.repository;

import com.park.utmstack.domain.UtmClient;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.Optional;


/**
 * Spring Data  repository for the UtmClient entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmClientRepository extends JpaRepository<UtmClient, Long> {
    Optional<UtmClient> findByClientName(String clientName);
}
