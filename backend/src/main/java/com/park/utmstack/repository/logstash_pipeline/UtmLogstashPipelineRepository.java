package com.park.utmstack.repository.logstash_pipeline;

import com.park.utmstack.domain.logstash_pipeline.UtmLogstashPipeline;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.stereotype.Repository;

import java.util.List;

/**
 * Spring Data SQL repository for the UtmLogstashPipeline entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmLogstashPipelineRepository extends JpaRepository<UtmLogstashPipeline, Long> {

    @Query("SELECT ulp FROM UtmLogstashPipeline ulp " +
        "LEFT JOIN UtmModule as um ON ulp.moduleName = um.moduleName " +
        "WHERE um.moduleActive IS NULL OR um.moduleActive = true")
    List<UtmLogstashPipeline> allActivePipelinesByServer();

    @Query(nativeQuery = true, value = "SELECT nextval('utm_logstash_pipeline_id_seq')")
    Long getNextId();

    @Query("SELECT distinct ulp.parentPipeline from UtmLogstashPipeline ulp " +
        "where ulp.parentPipeline is not null order by ulp.parentPipeline")
    List<Long> getParents();
}
