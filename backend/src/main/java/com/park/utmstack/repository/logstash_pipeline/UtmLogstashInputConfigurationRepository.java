package com.park.utmstack.repository.logstash_pipeline;

import com.park.utmstack.domain.logstash_pipeline.UtmLogstashInputConfiguration;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;

/**
 * Spring Data SQL repository for the UtmLogstashInputConfiguration entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmLogstashInputConfigurationRepository extends JpaRepository<UtmLogstashInputConfiguration, Long> {
    @Query(nativeQuery = true, value = "SELECT nextval('utm_logstash_input_configuration_id_seq')")
    Long getNextId();

    @Query("select ulic from UtmLogstashInputConfiguration ulic where ulic.confType in (:types)")
    List<UtmLogstashInputConfiguration> allConfigsByType(@Param("types") List<String> types);

    List<UtmLogstashInputConfiguration> getUtmLogstashInputConfigurationsByInputId(@Param("inputId") Integer inputId);
}
