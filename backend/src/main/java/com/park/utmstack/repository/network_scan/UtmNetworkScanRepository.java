package com.park.utmstack.repository.network_scan;

import com.park.utmstack.domain.network_scan.UtmNetworkScan;
import com.park.utmstack.domain.network_scan.enums.AssetRegisteredMode;
import com.park.utmstack.domain.network_scan.enums.AssetStatus;
import com.park.utmstack.domain.network_scan.enums.UpdateLevel;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.data.jpa.repository.Modifying;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.time.Instant;
import java.util.List;
import java.util.Optional;


/**
 * Spring Data  repository for the UtmNetworkScan entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmNetworkScanRepository extends JpaRepository<UtmNetworkScan, Long>, JpaSpecificationExecutor<UtmNetworkScan> {

    @Query(value = "SELECT DISTINCT ns FROM UtmNetworkScan ns " +
            "LEFT JOIN ns.dataInputIpList dti " +
            "LEFT JOIN ns.dataInputSourceList dts " +
            "WHERE" +
        "(:assetIpMacName IS NULL OR (ns.assetIp LIKE :assetIpMacName OR lower(ns.assetMac) LIKE lower(:assetIpMacName) OR lower(ns.assetName) LIKE lower(:assetIpMacName))) " +
        "AND ((:assetOs) IS NULL OR ns.assetOs IN :assetOs) " +
        "AND ((:assetType) IS NULL OR ns.assetTypeId IN (SELECT types.id FROM UtmAssetTypes types WHERE types.typeName IN :assetType)) " +
        "AND ((:groups) IS NULL OR ns.groupId IN (SELECT group.id FROM UtmAssetGroup group WHERE group.groupName IN :groups)) " +
        "AND ((:assetAlive) IS NULL OR ns.assetAlive IN :assetAlive) " +
        "AND ((:assetStatus) IS NULL OR ns.assetStatus IN :assetStatus) " +
        "AND ((:registeredMode) IS NULL OR ns.registeredMode = :registeredMode) " +
        "AND ((:assetAlias) IS NULL OR ns.assetAlias IN :assetAlias) " +
        "AND ((:serverName) IS NULL OR ns.serverName IN :serverName) " +
        "AND ((:isAgent) IS NULL OR ns.isAgent IN :isAgent) " +
        "AND ((:assetOsPlatform) IS NULL OR ns.assetOsPlatform IN :assetOsPlatform) " +
        "AND ((cast(:initDate as timestamp) is null) or (cast(:endDate as timestamp) is null) or (ns.discoveredAt BETWEEN :initDate AND :endDate)) " +
        "AND ((:dataTypes) IS NULL OR dti.dataType IN (:dataTypes) OR dts.dataType IN (:dataTypes))" +
        "AND ((:ports) IS NULL OR ns.id IN (SELECT DISTINCT ins.id FROM UtmNetworkScan ins INNER JOIN UtmPorts p ON ins.id = p.scanId WHERE p.port IN :ports))")
    Page<UtmNetworkScan> searchByFilters(@Param("assetIpMacName") String assetIpMacName,
                                         @Param("assetOs") List<String> assetOs,
                                         @Param("assetAlias") List<String> assetAlias,
                                         @Param("assetType") List<String> assetType,
                                         @Param("assetAlive") List<Boolean> assetAlive,
                                         @Param("assetStatus") List<AssetStatus> assetStatus,
                                         @Param("serverName") List<String> serverName,
                                         @Param("ports") List<Integer> ports,
                                         @Param("initDate") Instant initDate,
                                         @Param("endDate") Instant endDate,
                                         @Param("groups") List<String> groups,
                                         @Param("registeredMode") AssetRegisteredMode registeredMode,
                                         @Param("isAgent") List<Boolean> isAgent,
                                         @Param("assetOsPlatform") List<String> assetOsPlatform,
                                         @Param("dataTypes") List<String> dataTypes,
                                         Pageable pageable);

    @Modifying
    @Query("UPDATE UtmNetworkScan s SET s.assetTypeId = :assetTypeId WHERE s.id IN :assetIds")
    void updateType(@Param("assetIds") List<Long> assetIds,
                    @Param("assetTypeId") Long assetTypeId);

    @Modifying
    @Query("UPDATE UtmNetworkScan s SET s.groupId = :assetGroupId WHERE s.id IN :assetIds")
    void updateGroup(@Param("assetIds") List<Long> assetIds,
                     @Param("assetGroupId") Long assetGroupId);

    Optional<List<UtmNetworkScan>> findAllByAssetStatus(AssetStatus status);

    Optional<List<UtmNetworkScan>> findAllByServerNameAndRegisteredMode(String serverName, AssetRegisteredMode registeredMode);

    Optional<List<UtmNetworkScan>> findAllByGroupId(Long groupId);

    Optional<List<UtmNetworkScan>> findAllByRegisteredModeIsNot(AssetRegisteredMode registeredMode);

    List<UtmNetworkScan> findAllByRegisteredMode(AssetRegisteredMode registeredMode);

    List<UtmNetworkScan> findAllByIsAgentTrue();

    @Query("select n from UtmNetworkScan n where lower(n.assetName) = lower(:assetName)")
    Optional<UtmNetworkScan> findByAssetName(@Param("assetName") String assetName);

    List<UtmNetworkScan> findAllByUpdateLevelIsNullOrUpdateLevelIn(List<UpdateLevel> updateLevels);

    @Query("select n from UtmNetworkScan n where n.assetName = :assetNameOrIp or n.assetIp = :assetNameOrIp")
    Optional<UtmNetworkScan> findByNameOrIp(@Param("assetNameOrIp") String assetNameOrIp);

    @Modifying
    @Query(nativeQuery = true, value = "with sources as (select ds.source from utm_data_input_status ds where ds.data_type in :types)" +
        " delete from utm_network_scan asset where asset.asset_ip in (select src.source from sources src) " +
        "or asset.asset_name in (select src.source from sources src)")
    void deleteAllAssetsByDataType(@Param("types") List<String> types);

    @Query("select distinct n.assetOsPlatform from UtmNetworkScan n where n.assetOsPlatform is not null and n.isAgent is true")
    List<String> findAgentsOsPlatform();

    @Query(nativeQuery = true, value = "select n.asset_name from utm_network_scan n where n.asset_name is not null and n.is_agent is true and n.asset_alive is true and n.asset_status <> 'MISSING' and n.asset_os_platform = :platform")
    List<String> findAgentNamesByPlatform(@Param("platform") String platform);
}
