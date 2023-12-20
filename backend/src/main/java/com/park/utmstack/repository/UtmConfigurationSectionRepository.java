package com.park.utmstack.repository;

import com.park.utmstack.domain.UtmConfigurationSection;
import com.park.utmstack.domain.application_modules.enums.ModuleName;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.stereotype.Repository;

import java.util.List;


/**
 * Spring Data  repository for the UtmConfigurationSection entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmConfigurationSectionRepository extends JpaRepository<UtmConfigurationSection, Long>, JpaSpecificationExecutor<UtmConfigurationSection> {

    List<UtmConfigurationSection> findAllByModuleNameShort(ModuleName shortName);

    UtmConfigurationSection findByModuleNameShortAndSection(ModuleName module, String section);

}
