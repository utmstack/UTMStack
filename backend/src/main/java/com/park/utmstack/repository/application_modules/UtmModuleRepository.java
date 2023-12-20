package com.park.utmstack.repository.application_modules;

import com.park.utmstack.domain.application_modules.UtmModule;
import com.park.utmstack.domain.application_modules.enums.ModuleName;
import org.springframework.data.jpa.repository.EntityGraph;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;


/**
 * Spring Data  repository for the UtmModule entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmModuleRepository extends JpaRepository<UtmModule, Long>, JpaSpecificationExecutor<UtmModule> {

    @EntityGraph(attributePaths = {"moduleGroups", "moduleGroups.moduleGroupConfigurations"})
    UtmModule findByServerIdAndModuleName(Long serverId, ModuleName shortName);

    Integer countAllByModuleNameAndModuleActiveIsTrue(ModuleName shortName);

    @Query("select distinct m.moduleCategory from UtmModule m where m.moduleCategory is not null and (:serverId is null or m.serverId = :serverId)")
    List<String> findModuleCategories(@Param("serverId") Long serverId);
}
