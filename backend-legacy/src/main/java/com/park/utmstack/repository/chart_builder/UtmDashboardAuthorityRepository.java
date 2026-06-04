package com.park.utmstack.repository.chart_builder;

import com.park.utmstack.domain.chart_builder.UtmDashboardAuthority;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;


/**
 * Spring Data  repository for the UtmDashboardAuthority entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmDashboardAuthorityRepository extends JpaRepository<UtmDashboardAuthority, Long> {

}
