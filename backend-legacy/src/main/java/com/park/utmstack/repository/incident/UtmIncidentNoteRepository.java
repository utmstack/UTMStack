package com.park.utmstack.repository.incident;

import com.park.utmstack.domain.incident.UtmIncidentNote;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.stereotype.Repository;


/**
 * Spring Data  repository for the UtmIncidentNote entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmIncidentNoteRepository extends JpaRepository<UtmIncidentNote, Long>, JpaSpecificationExecutor<UtmIncidentNote> {

}
