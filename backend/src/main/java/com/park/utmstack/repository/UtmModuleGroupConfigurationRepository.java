package com.park.utmstack.repository;

import com.park.utmstack.domain.application_modules.UtmModuleGroupConfiguration;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;


/**
 * Spring Data  repository for the UtmModuleGroupConfiguration entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmModuleGroupConfigurationRepository extends JpaRepository<UtmModuleGroupConfiguration, Long> {

    List<UtmModuleGroupConfiguration> findAllByGroupId(Long groupId);

    UtmModuleGroupConfiguration findByGroupIdAndConfKey(Long groupId, String confKey);
}
