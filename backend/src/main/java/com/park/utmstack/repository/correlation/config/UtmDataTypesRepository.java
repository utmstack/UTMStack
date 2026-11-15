package com.park.utmstack.repository.correlation.config;

import com.park.utmstack.domain.correlation.config.UtmDataTypes;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;

import java.util.List;
import java.util.Optional;

public interface UtmDataTypesRepository extends JpaRepository<UtmDataTypes, Long>, JpaSpecificationExecutor<UtmDataTypes> {
    @Query(nativeQuery = true, value = "SELECT nextval('utm_data_types_id_seq')")
    Long getNextId();

    // Hibernate 7 sends a bare nullable JPQL String param to PostgreSQL as an untyped "?",
    // which fails with "could not determine data type of parameter $1" (on older Hibernate the
    // same pattern typed the param from context and failed with "function lower(bytea) does
    // not exist"). An explicit cast(:search as string) types it as varchar; a null cast stays
    // null, so the :search IS NULL guard behaves exactly as before. The JOIN FETCH on the
    // lazy ManyToOne module field is required because open-in-view is disabled and Jackson
    // cannot initialize a proxy outside the session (ManyToOne = no row duplication, so
    // pagination is unaffected).
    @Query(value = "SELECT dt FROM UtmDataTypes dt " +
            "LEFT JOIN FETCH dt.module WHERE" +
            "(:search IS NULL OR ((dt.dataType LIKE cast(:search as string) OR lower(dt.dataType) LIKE lower(cast(:search as string))) " +
            "OR (dt.dataTypeName LIKE cast(:search as string) OR lower(dt.dataTypeName) LIKE lower(cast(:search as string)))" +
            "OR (dt.dataTypeDescription LIKE cast(:search as string) OR lower(dt.dataTypeDescription) LIKE lower(cast(:search as string)))))")
    Page<UtmDataTypes> searchByFilters(@Param("search") String search,
                                       Pageable pageable);

    Optional<UtmDataTypes> findOneByDataType(String dataType);

    List<UtmDataTypes> findAllByIncludedFalse();

    @Query("select dt from UtmDataTypes dt where dt.systemOwner = false and dt.dataType not in (select ds.dataType from UtmDataInputStatus ds)")
    List<UtmDataTypes> findOrphanDataSourceConfigurations();

}
