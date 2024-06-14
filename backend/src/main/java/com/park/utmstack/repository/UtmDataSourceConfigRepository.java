package com.park.utmstack.repository;

import com.park.utmstack.domain.UtmDataSourceConfig;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

/**
 * Spring Data SQL repository for the UtmDataSourceConfig entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmDataSourceConfigRepository extends JpaRepository<UtmDataSourceConfig, UUID> {
    Optional<UtmDataSourceConfig> findOneByDataType(String dataType);

    List<UtmDataSourceConfig> findAllByIncludedFalse();

    @Query("select dsCfg from UtmDataSourceConfig dsCfg where dsCfg.systemOwner is false and dsCfg.dataType not in (select ds.dataType from UtmDataInputStatus ds)")
    List<UtmDataSourceConfig> findOrphanDataSourceConfigurations();


}
