package com.park.utmstack.repository.index_pattern;

import com.park.utmstack.domain.index_pattern.UtmIndexPattern;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;


/**
 * Spring Data  repository for the UtmIndexPattern entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmIndexPatternRepository extends JpaRepository<UtmIndexPattern, Long>, JpaSpecificationExecutor<UtmIndexPattern> {

    void deleteAllByPatternSystemIsTrueAndIdNotIn(List<Long> ids);

    Optional<UtmIndexPattern> findFirstByPatternSystemIsTrueOrderByIdDesc();

    @Query(nativeQuery = true, value = "select utm_index_pattern.* from utm_index_pattern where :nameShort = any(string_to_array(utm_index_pattern.pattern_module, ','))")
    List<UtmIndexPattern> findAllByPatternModule(@Param("nameShort") String nameShort);

}
