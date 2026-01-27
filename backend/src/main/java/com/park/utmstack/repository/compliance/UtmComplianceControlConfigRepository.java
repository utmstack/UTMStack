package com.park.utmstack.repository.compliance;

import com.park.utmstack.domain.compliance.UtmComplianceControlConfig;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.Optional;

@Repository
public interface UtmComplianceControlConfigRepository extends JpaRepository<UtmComplianceControlConfig, Long>, JpaSpecificationExecutor<UtmComplianceControlConfig> {
    @Query("""
        SELECT c FROM UtmComplianceControlConfig c
        LEFT JOIN FETCH c.section s
        LEFT JOIN FETCH s.standard st
        LEFT JOIN FETCH c.queriesConfigs q
        WHERE c.id = :id
    """)
    Optional<UtmComplianceControlConfig> findByIdWithQueries(@Param("id") Long id);
}
