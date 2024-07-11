package com.park.utmstack.repository.correlation.config;

import com.park.utmstack.domain.correlation.config.UtmRegexPattern;
import com.park.utmstack.domain.correlation.config.UtmTenantConfig;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;


public interface UtmTenantConfigRepository extends JpaRepository<UtmTenantConfig, Long>, JpaSpecificationExecutor<UtmTenantConfig> {
    @Query(value = "SELECT t FROM UtmTenantConfig t WHERE" +
            "(:search IS NULL OR ((t.assetName LIKE :search OR lower(t.assetName) LIKE lower(:search))))")
    Page<UtmTenantConfig> searchByFilters(@Param("search") String search,
                                          Pageable pageable);
}
