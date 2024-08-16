package com.park.utmstack.repository;

import com.park.utmstack.domain.application_modules.UtmModuleGroup;
import org.springframework.data.jpa.repository.EntityGraph;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;


/**
 * Spring Data  repository for the UtmConfigurationGroup entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmModuleGroupRepository extends JpaRepository<UtmModuleGroup, Long> {

    @EntityGraph(attributePaths = "moduleGroupConfigurations")
    List<UtmModuleGroup> findAllByModuleId(Long moduleId);

    @EntityGraph(attributePaths = "moduleGroupConfigurations")
    List<UtmModuleGroup> findAllByCollector(String collector);

    @EntityGraph(attributePaths = "moduleGroupConfigurations")
    List<UtmModuleGroup> findAllByCollectorIsNotNull();

    @EntityGraph(attributePaths = "moduleGroupConfigurations")
    List<UtmModuleGroup> findAllByModuleIdAndCollector(Long moduleId, String collector);

    void deleteAllByModuleId(Long moduleId);

    void deleteAllByCollector(String id);
}
