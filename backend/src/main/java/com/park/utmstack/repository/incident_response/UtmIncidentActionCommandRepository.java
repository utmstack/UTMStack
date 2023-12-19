package com.park.utmstack.repository.incident_response;

import com.park.utmstack.domain.incident_response.UtmIncidentActionCommand;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.Optional;


/**
 * Spring Data  repository for the UtmIncidentActionCommand entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmIncidentActionCommandRepository extends JpaRepository<UtmIncidentActionCommand, Long>, JpaSpecificationExecutor<UtmIncidentActionCommand> {

    @Query("select cmd.command from UtmIncidentActionCommand cmd where cmd.actionId = :actionId and cmd.osPlatform = :osPlatform")
    Optional<String> findCommand(@Param("osPlatform") String osPlatform, @Param("actionId") Long actionId);
}
