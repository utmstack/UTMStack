package com.utmstack.userauditor.repository;

import com.utmstack.userauditor.model.SourceScan;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;

/**
 * Spring Data SQL repository for the UtmAuditorUserSources entity.
 */

@Repository
public interface SourceScanRepository extends JpaRepository<SourceScan, Long> {
    List<SourceScan> findBySource_Id(Long id);
}