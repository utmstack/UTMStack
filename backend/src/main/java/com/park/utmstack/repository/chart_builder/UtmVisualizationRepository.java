package com.park.utmstack.repository.chart_builder;

import com.park.utmstack.domain.chart_builder.UtmVisualization;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;


/**
 * Spring Data  repository for the UtmVisualization entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmVisualizationRepository extends JpaRepository<UtmVisualization, Long>, JpaSpecificationExecutor<UtmVisualization> {

    void deleteByIdIn(List<Long> ids);

    void deleteAllBySystemOwnerIsTrueAndIdNotIn(List<Long> ids);

    Optional<UtmVisualization> findByName(String name);

    Optional<UtmVisualization> findByIdAndName(Long id, String name);

    Optional<UtmVisualization> findFirstBySystemOwnerIsTrueOrderByIdDesc();

    List<UtmVisualization> findAllByChTypeIn(List<String> types);
}
