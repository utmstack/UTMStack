package com.park.utmstack.repository;

import com.park.utmstack.domain.UtmSchedule;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;


/**
 * Spring Data  repository for the UtmSchedule entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmScheduleRepository extends JpaRepository<UtmSchedule, Long> {

}
