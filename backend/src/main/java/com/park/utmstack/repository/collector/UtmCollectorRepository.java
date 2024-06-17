package com.park.utmstack.repository.collector;

import com.park.utmstack.domain.collector.UtmCollector;
import com.park.utmstack.domain.network_scan.enums.AssetRegisteredMode;
import com.park.utmstack.domain.network_scan.enums.AssetStatus;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.data.jpa.repository.Modifying;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;

import java.time.Instant;
import java.util.List;
import java.util.Optional;

public interface UtmCollectorRepository extends JpaRepository<UtmCollector, Long>, JpaSpecificationExecutor<UtmCollector> {

    @Query(value = "SELECT ns FROM UtmCollector ns WHERE" +
            "(:assetIpMacName IS NULL OR (ns.ip LIKE :assetIpMacName OR lower(ns.hostname) LIKE lower(:assetIpMacName))) " +
            "AND ((:groups) IS NULL OR ns.groupId IN (SELECT group.id FROM UtmAssetGroup group WHERE group.groupName IN :groups)) " +
            "AND ((cast(:initDate as timestamp) is null) or (cast(:endDate as timestamp) is null) or (ns.lastSeen BETWEEN :initDate AND :endDate)) ")
    Page<UtmCollector> searchByFilters(@Param("assetIpMacName") String assetIpMacName,
                                         @Param("initDate") Instant initDate,
                                         @Param("endDate") Instant endDate,
                                         @Param("groups") List<String> groups,
                                         Pageable pageable);
    @Modifying
    @Query("UPDATE UtmCollector s SET s.groupId = :assetGroupId WHERE s.id IN :collectorsIds")
    void updateGroup(@Param("collectorsIds") List<Long> collectorsIds,
                     @Param("assetGroupId") Long assetGroupId);

    Optional<List<UtmCollector>> findAllByGroupId(Long groupId);
}
