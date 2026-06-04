package com.park.utmstack.repository.chart_builder;

import com.park.utmstack.domain.chart_builder.UtmDashboardVisualization;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;


/**
 * Spring Data  repository for the UtmDashboardVisualization entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmDashboardVisualizationRepository extends JpaRepository<UtmDashboardVisualization, Long>, JpaSpecificationExecutor<UtmDashboardVisualization> {

    Optional<List<UtmDashboardVisualization>> findAllByIdDashboard(Long idDashboard);

    Optional<UtmDashboardVisualization> findByIdDashboardAndIdVisualization(Long idDashboard, Long idVisualization);
}
