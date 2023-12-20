package com.park.utmstack.repository.incident_response;

import com.park.utmstack.domain.incident_response.UtmIncidentJob;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;


/**
 * Spring Data  repository for the UtmIncidentJob entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmIncidentJobRepository extends JpaRepository<UtmIncidentJob, Long>, JpaSpecificationExecutor<UtmIncidentJob> {
    List<UtmIncidentJob> findAllByAgent(String host);

    @Query("select job from UtmIncidentJob job where job.status = 0 and job.agent = :assetName order by job.id asc")
    List<UtmIncidentJob> findPendingCommands(@Param("assetName") String assetName);
}
