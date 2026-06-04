package com.park.utmstack.repository;

import com.park.utmstack.domain.UtmSpaceNotificationControl;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;


/**
 * Spring Data  repository for the UtmSpaceNotificationLast entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmSpaceNotificationControlRepository extends JpaRepository<UtmSpaceNotificationControl, Long> {

}
