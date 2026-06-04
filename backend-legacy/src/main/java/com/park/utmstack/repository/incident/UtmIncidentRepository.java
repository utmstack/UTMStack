package com.park.utmstack.repository.incident;

import com.park.utmstack.domain.incident.UtmIncident;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.stereotype.Repository;


/**
 * Spring Data  repository for the UtmIncident entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmIncidentRepository extends JpaRepository<UtmIncident, Long>, JpaSpecificationExecutor<UtmIncident> {

}
