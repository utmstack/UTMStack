package com.park.utmstack.repository.chart_builder;

import com.park.utmstack.domain.chart_builder.UtmDashboard;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;


/**
 * Spring Data  repository for the UtmDashboard entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmDashboardRepository extends JpaRepository<UtmDashboard, Long>, JpaSpecificationExecutor<UtmDashboard> {

    Optional<UtmDashboard> findByName(String name);

    Optional<UtmDashboard> findByIdAndName(Long id, String name);

    void deleteAllBySystemOwnerIsTrueAndIdNotIn(List<Long> ids);

    Optional<UtmDashboard> findFirstBySystemOwnerIsTrueOrderByIdDesc();

}
