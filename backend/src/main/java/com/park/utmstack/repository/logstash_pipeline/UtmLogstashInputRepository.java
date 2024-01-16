package com.park.utmstack.repository.logstash_pipeline;

import com.park.utmstack.domain.logstash_pipeline.UtmLogstashInput;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;

/**
 * Spring Data SQL repository for the UtmLogstashInput entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmLogstashInputRepository extends JpaRepository<UtmLogstashInput, Long> {
    @Query(nativeQuery = true, value = "SELECT nextval('utm_logstash_input_id_seq')")
    Long getNextId();

    List<UtmLogstashInput> getUtmLogstashInputsByPipelineId(@Param("pipelineId") Integer pipelineId);
}
