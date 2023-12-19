package com.park.utmstack.repository;

import com.park.utmstack.domain.UtmAlertLast;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;


/**
 * Spring Data  repository for the UtmAlertLast entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmAlertLastRepository extends JpaRepository<UtmAlertLast, Long> {

}
