package com.park.utmstack.repository.correlation.config;

import com.park.utmstack.domain.correlation.config.UtmRegexPattern;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.data.jpa.repository.Query;

public interface UtmRegexPatternRepository extends JpaRepository<UtmRegexPattern, Long>, JpaSpecificationExecutor<UtmRegexPattern> {
    @Query(nativeQuery = true, value = "SELECT nextval('utm_regex_pattern_id_seq')")
    Long getNextId();
}
