package com.park.utmstack.repository.correlation.config;

import com.park.utmstack.domain.correlation.config.UtmDataTypes;
import com.park.utmstack.domain.correlation.config.UtmRegexPattern;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;

public interface UtmRegexPatternRepository extends JpaRepository<UtmRegexPattern, Long>, JpaSpecificationExecutor<UtmRegexPattern> {
    @Query(nativeQuery = true, value = "SELECT nextval('utm_regex_pattern_id_seq')")
    Long getNextId();

    // Hibernate 7 sends bare nullable JPQL String params to PostgreSQL untyped
    // ("could not determine data type of parameter $1"); cast(:search as string) types
    // them as varchar. A null cast stays null, so the IS NULL guard behaves as before.
    @Query(value = "SELECT rp FROM UtmRegexPattern rp WHERE" +
            "(:search IS NULL OR ((rp.patternId LIKE cast(:search as string) OR lower(rp.patternId) LIKE lower(cast(:search as string))) " +
            "OR (rp.patternDescription LIKE cast(:search as string) OR lower(rp.patternDescription) LIKE lower(cast(:search as string)))))")
    Page<UtmRegexPattern> searchByFilters(@Param("search") String search,
                                       Pageable pageable);
}
