package com.park.utmstack.repository;

import com.park.utmstack.domain.UtmDataInputStatus;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;


/**
 * Spring Data  repository for the UtmDataInputStatus entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmDataInputStatusRepository extends JpaRepository<UtmDataInputStatus, String>, JpaSpecificationExecutor<UtmDataInputStatus> {

    List<UtmDataInputStatus> findAllByDataTypeIn(List<String> types);

    @Query("select u from UtmDataInputStatus u where u.dataType not in :excludedTypes and u.source in " +
        "(select u2.source from UtmDataInputStatus u2 where u2.source not in ('unknown') group by u2.source)")
    List<UtmDataInputStatus> extractSourcesToExport(@Param("excludedTypes") List<String> excludedTypes);

    @Query(value = "select d3 from UtmDataInputStatus d3 where d3.source in (select d.source from UtmDataInputStatus d left join UtmNetworkScan n on d.source = n.assetIp or d.source = n.assetName " +
        "left join UtmAssetTypes t on n.assetTypeId = t.id where (n.assetTypeId is null or t.typeName <> 'WORKSTATION') " +
        "and d.dataType not in ('generic', 'hids', 'utmstack') and d.source in (select u2.source from UtmDataInputStatus u2 group by u2.source))")
    List<UtmDataInputStatus> findDatasourceToCheckIfDown();

    @Query(value = "select d2 from UtmDataInputStatus d2 where d2.source in " +
        "(select d.source from UtmDataInputStatus d where d.dataType not in ('hids', 'utmstack', 'UTMStack') " +
        "and d.source in (select u2.source from UtmDataInputStatus u2 where u2.source not in ('unknown') group by u2.source))")
    Page<UtmDataInputStatus> findImportantDatasource(Pageable pageable);

    Optional<UtmDataInputStatus> findFirstByDataType(String dataType);

    @Query("select u from UtmDataInputStatus u where u.dataType in :excludedTypes")
    List<UtmDataInputStatus> extractExcludedSources(@Param("excludedTypes") List<String> excludedTypes);

    void deleteAllBySource(String source);

    /**
     * Extract data sources that are not already configured
     * @return A list of ${@link UtmDataInputStatus}
     */
    @Query("select distinct ds.dataType from UtmDataInputStatus ds where ds.dataType not in (select dt.dataType from UtmDataTypes dt) and ds.dataType != :dataType")
    List<String> findDataSourcesToConfigure(@Param("dataType") String dataType);
}
