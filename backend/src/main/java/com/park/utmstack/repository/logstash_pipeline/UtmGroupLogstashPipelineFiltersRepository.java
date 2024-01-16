package com.park.utmstack.repository.logstash_pipeline;

import com.park.utmstack.domain.logstash_pipeline.UtmGroupLogstashPipelineFilters;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Modifying;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;

/**
 * Spring Data SQL repository for the UtmGroupLogstashPipelineFilters entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmGroupLogstashPipelineFiltersRepository extends JpaRepository<UtmGroupLogstashPipelineFilters, Long> {

    @Query("select filters from UtmGroupLogstashPipelineFilters filters where filters.pipelineId = :pipelineId")
    List<UtmGroupLogstashPipelineFilters> getFilters(@Param("pipelineId") Integer pipelineId);

    @Modifying
    @Query("delete from UtmGroupLogstashPipelineFilters filters where filters.filterId = :filterId")
    void deleteRelationByFilterId(@Param("filterId") Integer filterId);
}
