package com.park.utmstack.repository.collector;

import com.park.utmstack.domain.collector.UtmCollector;
import com.park.utmstack.domain.network_scan.UtmNetworkScan;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.data.jpa.repository.Modifying;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;

import java.util.List;

public interface UtmCollectorRepository extends JpaRepository<UtmCollector, Long>, JpaSpecificationExecutor<UtmCollector> {
    @Modifying
    @Query("UPDATE UtmCollector s SET s.groupId = :assetGroupId WHERE s.id IN :collectorsIds")
    void updateGroup(@Param("collectorsIds") List<Long> collectorsIds,
                     @Param("assetGroupId") Long assetGroupId);
}
