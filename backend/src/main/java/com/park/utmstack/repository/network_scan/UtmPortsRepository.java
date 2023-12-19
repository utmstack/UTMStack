package com.park.utmstack.repository.network_scan;

import com.park.utmstack.domain.network_scan.UtmPorts;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.stereotype.Repository;

import java.util.List;


/**
 * Spring Data  repository for the UtmOpenPort entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmPortsRepository extends JpaRepository<UtmPorts, Long>, JpaSpecificationExecutor<UtmPorts> {
    void deleteAllByScanId(Long scanId);

    void deleteAllByScanIdIn(List<Long> scanIds);
}
