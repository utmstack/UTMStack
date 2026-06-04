package com.park.utmstack.repository.incident;

import com.park.utmstack.domain.incident.UtmIncidentHistory;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.stereotype.Repository;


/**
 * Spring Data  repository for the UtmIncidentHistory entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmIncidentHistoryRepository extends JpaRepository<UtmIncidentHistory, Long>, JpaSpecificationExecutor<UtmIncidentHistory> {

}
