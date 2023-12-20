package com.park.utmstack.repository.incident;

import com.park.utmstack.domain.incident.UtmIncidentAlert;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.data.jpa.repository.Modifying;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;


/**
 * Spring Data  repository for the UtmIncidentAlert entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmIncidentAlertRepository extends JpaRepository<UtmIncidentAlert, Long>, JpaSpecificationExecutor<UtmIncidentAlert> {

    @Modifying
    @Query(value = "update utm_incident_alert set  alert_status = :status where alert_id in :alertIds", nativeQuery = true)
    void updateAlertStatusByAlertIdIn(@Param("alertIds") List<String> alertIds, @Param("status") Integer status);

    List<UtmIncidentAlert> findAllByIncidentId(Long incidentId);
}
