package com.park.utmstack.repository.correlation.config;

import com.park.utmstack.domain.correlation.config.UtmRegexPattern;
import com.park.utmstack.domain.correlation.config.UtmTenantConfig;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;

import java.util.Optional;


public interface UtmTenantConfigRepository extends JpaRepository<UtmTenantConfig, Long>, JpaSpecificationExecutor<UtmTenantConfig> {

    // Hibernate 7 sends bare nullable JPQL String params to PostgreSQL untyped
    // ("could not determine data type of parameter $1"); cast(:search as string) types
    // them as varchar. A null cast stays null, so the IS NULL guard behaves as before.
    @Query(value = "SELECT t FROM UtmTenantConfig t WHERE" +
            "(:search IS NULL OR ((t.assetName LIKE cast(:search as string) OR lower(t.assetName) LIKE lower(cast(:search as string)))))")
    Page<UtmTenantConfig> searchByFilters(@Param("search") String search,
                                          Pageable pageable);

    Optional<UtmTenantConfig> findByAssetName(String assetName);
}
