package com.park.utmstack.repository.logstash_pipeline;

import com.park.utmstack.domain.logstash_pipeline.UtmPortsConfiguration;
import org.springframework.data.jpa.repository.*;
import org.springframework.stereotype.Repository;

/**
 * Spring Data SQL repository for the UtmPortsConfiguration entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmPortsConfigurationRepository extends JpaRepository<UtmPortsConfiguration, Long> {}
