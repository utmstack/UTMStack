package com.utmstack.userauditor.repository;

import com.utmstack.userauditor.model.SourceScan;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.Optional;

/**
 * Spring Data SQL repository for the UtmAuditorUserSources entity.
 */

@Repository
public interface SourceScanRepository extends JpaRepository<SourceScan, Long> {

    @Query("SELECT ss FROM SourceScan ss WHERE ss.source.id = :sourceId ORDER BY ss.executionDate DESC")
    Optional<SourceScan> findLatestBySourceId(@Param("sourceId") Long sourceId);
}