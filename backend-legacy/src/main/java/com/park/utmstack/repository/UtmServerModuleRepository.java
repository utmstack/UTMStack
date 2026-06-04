package com.park.utmstack.repository;

import com.park.utmstack.domain.UtmServerModule;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;


/**
 * Spring Data  repository for the UtmServerModule entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmServerModuleRepository extends JpaRepository<UtmServerModule, Long>, JpaSpecificationExecutor<UtmServerModule> {

    @Query(value = "select m from UtmServerModule m where m.id in (select i.moduleId from UtmIntegration i where i.moduleId is not null) and (:serverId is null or m.serverId = :serverId) and (:prettyName is null or m.prettyName like :prettyName) order by m.id")
    List<UtmServerModule> getModulesWithIntegrations(@Param("serverId") Long serverId, @Param("prettyName") String prettyName);

    List<UtmServerModule> findAllByModuleName(String moduleName);
}
